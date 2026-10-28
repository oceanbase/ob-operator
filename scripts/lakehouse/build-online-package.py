#!/usr/bin/env python3
"""Build a small internal online bundle from a validated offline image profile.

No image payload, user configuration, state, or credentials enter the output.
Image archives are fetched on demand from explicit, checksum-pinned URLs.
"""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import re
import shutil
import subprocess
from urllib.parse import urlsplit

SPEC = importlib.util.spec_from_file_location("offline_builder", Path(__file__).with_name("build-package.py"))
BASE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(BASE)


def build(args):
    url = urlsplit(args.archive_url)
    if url.scheme not in ("http", "https") or not url.hostname or url.username or url.password or url.query or url.fragment:
        raise ValueError("archive-url requires HTTP(S), without credentials, query or fragment")
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]{0,80}", args.version):
        raise ValueError("Invalid version")
    sha, timestamp = BASE.source_identity(args.source, args.source_archive)
    if args.output.exists():
        raise ValueError("Output directory already exists")
    for path in (args.installer, args.helm, args.jq):
        header = path.read_bytes()[:20]
        if header[:5] != b"\x7fELF\x02" or int.from_bytes(header[18:20], "little") != 62:
            raise ValueError("Expected a Linux AMD64 ELF binary: " + str(path))
    args.output.mkdir(parents=True)
    name = "oceanbase-lakehouse-online-%s-linux-amd64" % args.version
    bundle = args.output / name
    bundle.mkdir()
    for folder in ("charts", "values", "prerequisites"):
        shutil.copytree(args.baseline / folder, bundle / folder)
    shutil.copyfile(args.source / "LICENSE", bundle / "LICENSE")
    shutil.copyfile(args.source / "install.sh", bundle / "install.sh")
    (bundle / "install.sh").chmod(0o755)
    for destination, source in (("bin/lakehouse-installer", args.installer), ("tools/bin/helm", args.helm), ("tools/bin/jq", args.jq)):
        target = bundle / destination
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, target)
        target.chmod(0o755)
    (bundle / "tools/bin/kubectl").write_text('''#!/usr/bin/env bash
set -euo pipefail
# Reuse a real local client; K3s already includes kubectl.
IFS=: read -r -a client_dirs <<< "$PATH"
for directory in "${client_dirs[@]}"; do
  client="${directory:-.}/kubectl"
  if [[ -f "$client" && -x "$client" && ! "$client" -ef "${BASH_SOURCE[0]}" ]]; then exec "$client" "$@"; fi
done
if command -v k3s >/dev/null 2>&1; then exec k3s kubectl "$@"; fi
echo 'kubectl is missing. Prepare Kubernetes/K3s and its client before installing.' >&2
exit 127
''')
    (bundle / "tools/bin/kubectl").chmod(0o755)
    (bundle / "tools/licenses").mkdir()
    for name_, source in (("helm-LICENSE", args.helm_license), ("jq-COPYING", args.jq_license)):
        shutil.copyfile(source, bundle / "tools/licenses" / name_)
    manifest = json.loads((args.baseline / "images/archives.json").read_text())
    serving = {}
    for archive in manifest["archives"]:
        if archive["base"] != "bundle":
            raise ValueError("Baseline must be self-contained")
        source = args.baseline / archive["path"]
        if BASE.digest(source) != archive["sha256"]:
            raise ValueError("Baseline archive checksum mismatch: " + archive["name"])
        filename = archive["sha256"] + ".tar.gz"
        serving["/archives/" + filename] = str(source.resolve())
        archive.update(path="images/cache/" + filename,
                       url=args.archive_url.rstrip("/") + "/" + filename,
                       size_bytes=source.stat().st_size)
    (bundle / "images").mkdir()
    BASE.write_json(bundle / "images/archives.json", manifest)
    lock = json.loads((args.baseline / "image-lock.json").read_text())
    lock["delivery_mode"] = "online-verified-archives"
    lock["online_installer_source_sha"] = sha
    lock["online_version"] = args.version
    for image in lock["images"]:
        image["delivery"] = "online-verified-archive"
    BASE.write_json(bundle / "image-lock.json", lock)
    tools = {"helm": {"version": args.helm_version, "sha256": BASE.digest(args.helm)},
             "jq": {"version": args.jq_version, "sha256": BASE.digest(args.jq)},
             "kubectl": {"delivery": "shim using local kubectl or K3s"}}
    BASE.write_json(bundle / "tools/versions.json", tools)
    readme = (args.source / "scripts/lakehouse/ONLINE-INSTALL.txt").read_text()
    readme = readme.replace("ONLINE_BUNDLE_NAME", bundle.name).replace("ONLINE_ARCHIVE_NAME", bundle.name + ".tar.gz")
    (bundle / "README.txt").write_text(readme)
    # Only immutable distributable files enter the checksum manifest. Online
    # payloads and private installation state are created after extraction.
    entries = []
    for path in sorted(bundle.rglob("*")):
        if path.is_file():
            entries.append(BASE.digest(path) + "  " + path.relative_to(bundle).as_posix())
    (bundle / "SHA256SUMS").write_text("\n".join(entries) + "\n")
    archive = args.output / (bundle.name + ".tar.gz")
    BASE.tar_tree(bundle, archive, bundle.name, timestamp, compressed=True)
    record = {"package": str(archive), "sha256": BASE.digest(archive), "size_bytes": archive.stat().st_size,
              "source_sha": sha, "image_profile_source_sha": lock.get("source_sha"),
              "mode": "online-verified-archives", "image_payload_files": 0,
              "image_references": len(lock["images"]), "tools": tools,
              "archive_routes": serving}
    BASE.write_json(args.output / "build-result.json", record)
    print(json.dumps({k: v for k, v in record.items() if k != "archive_routes"}, indent=2))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("source", "baseline", "installer", "helm", "jq", "helm-license", "jq-license", "output"):
        parser.add_argument("--" + name, type=Path, required=True)
    parser.add_argument("--source-archive", type=Path)
    for name in ("archive-url", "version", "helm-version", "jq-version"):
        parser.add_argument("--" + name, required=True)
    build(parser.parse_args())

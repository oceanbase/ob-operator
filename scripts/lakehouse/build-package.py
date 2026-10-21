#!/usr/bin/env python3
"""Build a self-contained Linux/amd64 QA tar without rebuilding any image.

Build-host dependencies: Python 3.8+, PyYAML, Git and Helm. None of the Python
dependencies are required on installation machines. Use a clean source commit.
The output directory must not exist. Failed builds are retained for diagnosis.
"""
import argparse
import difflib
from datetime import datetime
import gzip
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import socket
import subprocess
import sys
import tarfile

try:
    import yaml
except ImportError:
    raise SystemExit("Build host requires PyYAML; use the isolated packaging environment")


CORE_ROLES = {"operator", "dashboard", "observer", "logservice"}
MINIO_ROLES = {"minio", "mc"}
DIAGNOSTIC_ROLES = {"obdiag", "oceanbase-helper", "observer", "dashboard"}
SHA = re.compile(r"^[0-9a-f]{64}$")


def run(args, cwd=None):
    result = subprocess.run(args, cwd=cwd, stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE, text=True, timeout=300)
    if result.returncode:
        # Commands here operate on build fixtures, never live Kubernetes data.
        raise RuntimeError("Build command failed: %s\n%s" %
                           (" ".join(map(str, args)), result.stderr.strip()))
    return result.stdout


def digest(path):
    result = hashlib.sha256()
    with Path(path).open("rb") as stream:
        for block in iter(lambda: stream.read(8 * 1024 * 1024), b""):
            result.update(block)
    return result.hexdigest()


def source_identity(source, export=None):
    if export is None:
        if run(["git", "status", "--porcelain"], cwd=source).strip():
            raise ValueError("Source checkout must be clean; commit the reviewed source before packaging")
        revision = run(["git", "rev-parse", "HEAD"], cwd=source).strip()
        timestamp = int(run(["git", "show", "-s", "--format=%ct", "HEAD"], cwd=source).strip())
        return revision, timestamp
    seen, timestamp = set(), None
    with tarfile.open(str(export), "r:*") as archive:
        revision = archive.pax_headers.get("comment", "").strip()
        if not re.fullmatch(r"[0-9a-f]{40}", revision):
            raise ValueError("Source export must be a git archive with its original commit PAX comment")
        for item in archive:
            relative = Path(item.name)
            if relative.is_absolute() or ".." in relative.parts:
                raise ValueError("Unsafe source archive path")
            path = source / relative
            if timestamp is None:
                timestamp = int(item.mtime)
            if item.isfile():
                if not path.is_file() or path.is_symlink():
                    raise ValueError("Source export file missing or replaced: " + str(relative))
                expected = hashlib.sha256()
                with archive.extractfile(item) as stream:
                    for block in iter(lambda: stream.read(1024 * 1024), b""):
                        expected.update(block)
                if expected.hexdigest() != digest(path):
                    raise ValueError("Source export differs from git archive: " + str(relative))
                seen.add(relative.as_posix())
            elif item.issym():
                if not path.is_symlink() or os.readlink(path) != item.linkname:
                    raise ValueError("Source export symlink differs: " + str(relative))
            elif not item.isdir():
                raise ValueError("Unsupported source archive member: " + str(relative))
    required = [source / n for n in ("install.sh", "INSTALL.txt", "LICENSE")]
    for chart in ("oceanbase-lakehouse", "oceanbase-dashboard", "ob-operator"):
        required.extend(p for p in (source / "charts" / chart).rglob("*") if p.is_file())
    for path in required:
        if path.relative_to(source).as_posix() not in seen:
            raise ValueError("Untracked input in exported source: " + str(path.relative_to(source)))
    if timestamp is None:
        raise ValueError("Source export is empty")
    return revision, timestamp


def write_json(path, value):
    Path(path).write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def write_yaml(path, value):
    Path(path).write_text(yaml.safe_dump(value, sort_keys=False, allow_unicode=True))


def normalize_ref(value):
    value = value[10:] if value.startswith("docker.io/") else value
    return value[8:] if value.startswith("library/") else value


def archive_image_ids(path):
    """Read actual Docker-save config bytes, not tag-only or manifest digests.

    Streaming traversal avoids retaining the large image layers in memory. Config
    JSON and manifest.json are the only payloads read into memory.
    """
    configs, manifests = {}, []
    with tarfile.open(str(path), "r|gz") as archive:
        for member in archive:
            name = member.name.lstrip("./")
            if not member.isfile():
                continue
            if name == "manifest.json":
                if member.size > 8 * 1024 * 1024:
                    raise ValueError("Unreasonably large Docker-save manifest")
                manifests.append(json.load(archive.extractfile(member)))
            elif member.size <= 16 * 1024 * 1024 and (name.endswith(".json") or name.startswith("blobs/sha256/")):
                # Docker 26's save format uses OCI-style blob paths, while
                # older save files name configs <sha>.json. Small layers are
                # discarded immediately instead of accumulating their bytes.
                payload = archive.extractfile(member).read()
                try:
                    metadata = json.loads(payload)
                except (ValueError, UnicodeDecodeError):
                    continue
                if isinstance(metadata, dict) and "architecture" in metadata and "os" in metadata:
                    configs[name] = {"id": "sha256:" + hashlib.sha256(payload).hexdigest(),
                                     "architecture": metadata["architecture"], "os": metadata["os"]}
    if len(manifests) != 1 or not isinstance(manifests[0], list):
        raise ValueError("Expected one Docker-save manifest.json in " + str(path))
    images = {}
    for entry in manifests[0]:
        config_name = entry["Config"].lstrip("./")
        if config_name not in configs:
            raise ValueError("Missing image config %s in %s" % (config_name, path))
        metadata = configs[config_name]
        if metadata.get("architecture") != "amd64" or metadata.get("os") != "linux":
            raise ValueError("Image archive is not Linux/amd64: " + config_name)
        image_id = metadata["id"]
        for reference in entry.get("RepoTags") or []:
            reference = normalize_ref(reference)
            if reference in images and images[reference] != image_id:
                raise ValueError("Conflicting image tag in archive: " + reference)
            images[reference] = image_id
    return images


def verify_archive(path, expected, expected_sha=None, exact=False):
    actual = archive_image_ids(path)
    problems = []
    for image in expected:
        observed = actual.get(normalize_ref(image["reference"]))
        if observed != image["image_id"]:
            problems.append("%s expected %s, got %s" %
                            (image["reference"], image["image_id"], observed))
    if problems:
        raise ValueError("Archive image ID mismatch in %s:\n%s" % (path, "\n".join(problems)))
    if exact:
        extra = actual.keys() - {normalize_ref(image["reference"]) for image in expected}
        if extra:
            raise ValueError("Untracked diagnostic archive images: " + ", ".join(sorted(extra)))
    archive_sha = digest(path)
    if expected_sha and archive_sha != expected_sha:
        raise ValueError("Baseline archive SHA256 mismatch: " + str(path))
    return archive_sha


def apply_diagnostic_profile(lock, profile_path):
    """Merge explicitly supplied, immutable diagnostic images into the lock.

    Base images remain covered because their bytes still occur in the unchanged
    core archive. Updated runtime roles must use distinct references; reusing a
    tag for a different ID would make import order determine the running image.
    """
    profile = json.loads(Path(profile_path).read_text())
    if profile.get("schemaVersion") != 1 or not SHA.fullmatch(profile.get("archive_sha256", "")):
        raise ValueError("Diagnostics lock requires schemaVersion 1 and archive_sha256")
    additions = profile.get("images", [])
    if not isinstance(additions, list) or any(not isinstance(entry, dict) for entry in additions):
        raise ValueError("Diagnostics images must be a list of image records")
    roles, refs = set(), {normalize_ref(entry["reference"]) for entry in lock["images"]}
    original = {entry["role"]: entry for entry in lock["images"]}
    for entry in additions:
        role, reference, image_id = entry.get("role"), entry.get("reference", ""), entry.get("image_id", "")
        if role not in DIAGNOSTIC_ROLES or role in roles:
            raise ValueError("Unexpected or duplicate diagnostics role: " + str(role))
        if not reference or normalize_ref(reference) in refs or any(ch.isspace() for ch in reference) or reference.startswith("-"):
            raise ValueError("Diagnostics images require distinct, valid references: " + reference)
        if reference.endswith(":latest") or (":" not in reference.rsplit("/", 1)[-1] and "@sha256:" not in reference):
            raise ValueError("Diagnostics images require an explicit non-latest version or digest: " + reference)
        if not re.fullmatch(r"sha256:[0-9a-f]{64}", image_id):
            raise ValueError("Missing immutable diagnostics config image ID: " + reference)
        if role == "observer" and entry.get("base_image_id") != original.get("observer", {}).get("image_id"):
            raise ValueError("Observer derivative base_image_id does not match the baseline Observer")
        if role in original and role + "-base" in original:
            raise ValueError("Use the original baseline, not a recursively updated diagnostics bundle")
        roles.add(role)
        refs.add(normalize_ref(reference))
    if not {"obdiag", "oceanbase-helper"} <= roles:
        raise ValueError("Diagnostics lock must include obdiag and oceanbase-helper")
    for entry in additions:
        role = entry["role"]
        if role in original:
            original[role]["role"] = role + "-base"
        merged = dict(entry)
        merged["delivery"] = "archive"
        lock["images"].append(merged)
    return profile


def configure_diagnostics(values, images):
    if "obdiag" not in images:
        return
    inspection = values.setdefault("inspection", {})
    for section, role in (("obdiag", "obdiag"), ("oceanbaseHelper", "oceanbase-helper")):
        inspection.setdefault(section, {})["image"] = images[role]["reference"]


def tar_tree(root, target, prefix, timestamp, compressed=False):
    """Canonical tar headers are produced by Python, independent of GNU tar."""
    root, target = Path(root), Path(target)
    with target.open("xb") as raw:
        gzip_stream = gzip.GzipFile(filename="", mode="wb", fileobj=raw, mtime=0, compresslevel=1) if compressed else None
        try:
            with tarfile.open(fileobj=gzip_stream or raw, mode="w|", format=tarfile.GNU_FORMAT) as archive:
                paths = [root] + sorted(root.rglob("*"), key=lambda p: str(p.relative_to(root)).encode())
                for path in paths:
                    if path.is_symlink() or not (path.is_file() or path.is_dir()):
                        raise ValueError("Unsupported source special file: " + str(path))
                    name = prefix if path == root else prefix + "/" + path.relative_to(root).as_posix()
                    item = archive.gettarinfo(str(path), arcname=name)
                    item.uid = item.gid = 0
                    item.uname = item.gname = ""
                    item.mtime = timestamp
                    item.mode = 0o755 if path.is_dir() or path.stat().st_mode & 0o111 else 0o644
                    if path.is_file():
                        with path.open("rb") as source:
                            archive.addfile(item, source)
                    else:
                        archive.addfile(item)
        finally:
            if gzip_stream:
                gzip_stream.close()


def copy_chart(source, destination):
    def ignore(directory, names):
        return [name for name in names if name in (".git", "__pycache__", "tests")
                or name.endswith(("_test.go", ".pyc", ".tgz"))]
    shutil.copytree(source, destination, ignore=ignore)


def patch_charts(source, package, images, timestamp):
    charts = package / "charts"
    charts.mkdir()
    for name in ("oceanbase-lakehouse", "oceanbase-dashboard", "ob-operator"):
        copy_chart(source / "charts" / name, charts / name)
    dashboard = charts / "oceanbase-dashboard"
    values = yaml.safe_load((dashboard / "values.yaml").read_text())
    values["image"] = images["dashboard"]["reference"]
    values["adminPassword"] = ""
    configure_diagnostics(values, images)
    write_yaml(dashboard / "values.yaml", values)
    credentials = dashboard / "templates/user-credentials.yaml"
    original = credentials.read_text()
    old = '(.Values.adminPassword | default ("admin"))'
    new = '(required "Pass adminPassword using --set-file from a protected password file" .Values.adminPassword)'
    if original.count(old) == 1:
        credentials.write_text(original.replace(old, new))
    elif new not in original:
        raise ValueError("Dashboard credential template changed; review the QA required-password patch")
    parent = charts / "oceanbase-lakehouse"
    values = yaml.safe_load((parent / "values.yaml").read_text())
    for section, role in (("dashboard", "dashboard"), ("logService", "logservice"),
                          ("observer", "observer"), ("minio", "minio"), ("initialization", "mc")):
        values[section]["image"] = images[role]["reference"]
    configure_diagnostics(values["dashboard"], images)
    write_yaml(parent / "values.yaml", values)
    operator = charts / "ob-operator/templates/operator.yaml"
    content, count = re.subn(r"(?m)^(\s*image:\s*)quay\.io/oceanbase/ob-operator:[^\s]+\s*$",
                             lambda m: m.group(1) + images["operator"]["reference"], operator.read_text())
    if count != 1:
        raise ValueError("Expected exactly one Operator image to replace, found %d" % count)
    operator.write_text(content)
    parent_meta = yaml.safe_load((parent / "Chart.yaml").read_text())
    child_meta = yaml.safe_load((dashboard / "Chart.yaml").read_text())
    dependencies = parent_meta.get("dependencies", [])
    if len(dependencies) != 1 or dependencies[0].get("repository") != "file://../oceanbase-dashboard" or str(dependencies[0].get("version")) != str(child_meta["version"]):
        raise ValueError("Unexpected lakehouse dependency; offline dependency requires review")
    (parent / "charts").mkdir(exist_ok=True)
    tar_tree(dashboard, parent / "charts" / ("oceanbase-dashboard-%s.tgz" % child_meta["version"]),
             "oceanbase-dashboard", timestamp, compressed=True)
    changes = []
    for name in ("oceanbase-lakehouse", "oceanbase-dashboard", "ob-operator"):
        for path in sorted((charts / name).rglob("*")):
            if not path.is_file() or path.suffix == ".tgz":
                continue
            relative = path.relative_to(charts)
            before = source / "charts" / relative
            if before.is_file() and before.read_bytes() != path.read_bytes():
                changes.extend(difflib.unified_diff(before.read_text().splitlines(True), path.read_text().splitlines(True),
                               fromfile="source/charts/" + str(relative), tofile="delivery/charts/" + str(relative)))
    (package / "provenance/qa-packaging.patch").write_text("".join(changes))
    return {"lakehouse": str(parent_meta["version"]), "dashboard": str(child_meta["version"]),
            "operator": str(yaml.safe_load((charts / "ob-operator/Chart.yaml").read_text())["version"])}


def quickstart_profile(baseline, package, images):
    profile = baseline / "values/single-worker-8core.yaml"
    if not profile.is_file():
        profile = baseline / "values/single-worker.yaml"
    values = yaml.safe_load(profile.read_text())
    values["storageClass"] = ""  # the Go preflight selects the sole default class
    values.setdefault("logService", {}).pop("clusterId", None)
    values.setdefault("observer", {}).pop("clusterId", None)
    values["logService"]["resource"]["cpu"] = "1"
    values["observer"]["resource"]["cpu"] = "2"
    for section, role in (("dashboard", "dashboard"), ("logService", "logservice"),
                          ("observer", "observer"), ("minio", "minio"), ("initialization", "mc")):
        values.setdefault(section, {})["image"] = images[role]["reference"]
    values["dashboard"]["enabled"] = True
    configure_diagnostics(values["dashboard"], images)
    # Old bundles used ClusterIP; the new single-worker entry advertises a
    # browser-ready NodePort. Let Kubernetes allocate it for each fresh release.
    service = values["dashboard"].setdefault("service", {})
    service["type"] = "NodePort"
    service.pop("nodePort", None)
    (package / "values").mkdir()
    write_yaml(package / "values/quickstart.yaml", values)
    if "obdiag" in images:
        # shared_storage keeps the Observer image immutable. A diagnostics
        # upgrade must work for existing clusters without changing that image.
        upgrade = {"dashboard": {"image": images["dashboard"]["reference"]}}
        configure_diagnostics(upgrade["dashboard"], images)
        write_yaml(package / "values/diagnostics-upgrade.yaml", upgrade)


def rendered_images(text):
    images = set()
    def visit(value):
        if isinstance(value, dict):
            # These images are launched by Dashboard jobs, so Helm renders them
            # inside config.yaml rather than as Pod container specifications.
            # SQL analyzer remains an explicitly excluded optional component.
            if value.get("kind") == "ConfigMap":
                configured = (value.get("data") or {}).get("config.yaml")
                if isinstance(configured, str):
                    config = yaml.safe_load(configured)
                    if isinstance(config, dict):
                        visit(config.get("inspection", {}))
            for key, item in value.items():
                if key == "image" and isinstance(item, str):
                    images.add(item)
                else:
                    visit(item)
        elif isinstance(value, list):
            for item in value:
                visit(item)
    for document in yaml.safe_load_all(text):
        visit(document)
    return images


def validate_charts(package, lock):
    charts = package / "charts"
    results = []
    for name in ("oceanbase-lakehouse", "oceanbase-dashboard", "ob-operator"):
        args = ["helm", "lint", "--strict", str(charts / name)]
        if name == "oceanbase-dashboard":
            args += ["--set-string", "adminPassword=render-only-not-an-account"]
        run(args)
        results.append("PASS strict lint " + name)
    references = set()
    for stage in ("storage", "logservice", "cluster"):
        for enabled in ("true", "false"):
            rendered = run(["helm", "template", "qa", str(charts / "oceanbase-lakehouse"), "-n", "qa",
                            "-f", str(package / "values/quickstart.yaml"), "--set-string", "storageClass=local-path",
                            "--set", "stage=" + stage, "--set", "dashboard.enabled=" + enabled])
            references |= rendered_images(rendered)
            results.append("PASS render stage=%s dashboard=%s" % (stage, enabled))
    references |= rendered_images(run(["helm", "template", "qa", str(charts / "ob-operator")]))
    references |= rendered_images(run(["helm", "template", "qa", str(charts / "oceanbase-dashboard"),
                                       "--set-string", "adminPassword=render-only-not-an-account"]))
    references |= rendered_images((package / "prerequisites/cert-manager-v1.12.4.yaml").read_text())
    locked = {item["reference"] for item in lock["images"]}
    missing = references - locked
    if missing:
        raise ValueError("Rendered images missing from image-lock.json: " + ", ".join(sorted(missing)))
    # Verify that standalone Dashboard never falls back to the upstream admin password.
    guard = subprocess.run(["helm", "template", "qa", str(charts / "oceanbase-dashboard")],
                           stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, text=True, timeout=60)
    if guard.returncode == 0 or "Pass adminPassword" not in guard.stderr:
        raise ValueError("Standalone Dashboard required-password guard did not reject an empty password")
    results += ["PASS standalone Dashboard password guard", "PASS rendered image closure (%d references)" % len(references)]
    return results


def build(args):
    source, baseline, output = args.source.resolve(), args.baseline.resolve(), args.output.resolve()
    minio_cache = args.minio_cache.resolve()
    installer = (args.installer_binary or source / "bin/lakehouse-installer").resolve()
    if output.exists():
        raise ValueError("Output already exists; choose a new directory: " + str(output))
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]{0,79}", args.version):
        raise ValueError("Version must contain 1-80 filename-safe characters")
    source_export = getattr(args, "source_archive", None)
    if source_export is not None:
        source_export = source_export.resolve()
    source_sha, timestamp = source_identity(source, source_export)
    with installer.open("rb") as binary:
        elf = binary.read(20)
    if len(elf) < 20 or elf[:6] != b"\x7fELF\x02\x01" or elf[18:20] != b"\x3e\x00":
        raise ValueError("Installer must be a compiled Linux/amd64 ELF binary; cross-compile cmd/lakehouse-installer first")
    lock = json.loads((baseline / "image-lock.json").read_text())
    baseline_sha = lock["source_sha"]
    diagnostics_archive = getattr(args, "diagnostics_archive", None)
    diagnostics_lock = getattr(args, "diagnostics_lock", None)
    if bool(diagnostics_archive) != bool(diagnostics_lock):
        raise ValueError("Pass both --diagnostics-archive and --diagnostics-lock")
    diagnostic_profile = None
    if diagnostics_lock:
        diagnostics_archive = diagnostics_archive.resolve()
        diagnostic_profile = apply_diagnostic_profile(lock, diagnostics_lock.resolve())
    baseline_checksums = {}
    for line in (baseline / "SHA256SUMS").read_text().splitlines():
        fields = line.split(None, 1)
        if len(fields) == 2 and SHA.fullmatch(fields[0]):
            path = fields[1].lstrip("*")
            baseline_checksums[path[2:] if path.startswith("./") else path] = fields[0]
    by_role = {}
    for image in lock["images"]:
        image_id = image.get("image_id", "")
        if image["role"] in by_role or not image_id.startswith("sha256:") or not SHA.fullmatch(image_id[7:]):
            raise ValueError("Duplicate role or invalid config image ID in baseline lock")
        by_role[image["role"]] = image
    if not (CORE_ROLES | MINIO_ROLES) <= by_role.keys():
        raise ValueError("Baseline is missing required Operator/Dashboard/DB/MinIO images")
    updated_roles = {entry["role"] for entry in diagnostic_profile["images"]} if diagnostic_profile else set()
    core_roles = (CORE_ROLES - updated_roles) | {role + "-base" for role in updated_roles & CORE_ROLES}
    input_archives = [
        ("core", baseline / "images/core-images.tar.gz", [i for i in lock["images"] if i["role"] in core_roles]),
        ("support", baseline / "images/support-images.tar.gz", [i for i in lock["images"] if i["role"] not in core_roles | MINIO_ROLES | updated_roles]),
        ("minio", minio_cache, [i for i in lock["images"] if i["role"] in MINIO_ROLES]),
    ]
    if diagnostic_profile:
        input_archives.append(("diagnostics", diagnostics_archive,
                               [i for i in lock["images"] if i["role"] in updated_roles]))
    for _, path, _ in input_archives:
        if not path.is_file():
            raise ValueError("Required delivery archive is missing: " + str(path))
    for filename in ("core-images.tar.gz", "support-images.tar.gz"):
        if "images/" + filename not in baseline_checksums:
            raise ValueError("Baseline SHA256SUMS does not cover images/" + filename)
    name = "oceanbase-lakehouse-qa-%s-linux-amd64" % args.version
    minio_name = "minio-images.tar.gz"
    package = output / name
    package.mkdir(parents=True)
    (package / "provenance").mkdir()
    (package / "images").mkdir()
    archive_entries, validation = [], []
    for role, path, expected in input_archives:
        print("Validate and reuse %s archive: %s" % (role, path), flush=True)
        expected_sha = baseline_checksums.get("images/" + path.name) if role in ("core", "support") else None
        if role == "diagnostics":
            expected_sha = diagnostic_profile["archive_sha256"]
        if role == "minio" and Path(str(path) + ".sha256").is_file():
            expected_sha = Path(str(path) + ".sha256").read_text().split()[0]
            if not SHA.fullmatch(expected_sha):
                raise ValueError("Invalid MinIO cache checksum sidecar")
        archive_sha = verify_archive(path, expected, expected_sha, exact=role == "diagnostics")
        filename = {"diagnostics": "diagnostics-images.tar.gz", "minio": minio_name}.get(role, path.name)
        destination = package / "images" / filename
        shutil.copyfile(path, destination)
        destination.chmod(0o644)
        if digest(destination) != archive_sha:
            raise ValueError("Archive copy checksum mismatch: " + str(destination))
        archive_entries.append({"name": role, "base": "bundle", "path": "images/" + filename,
                                "sha256": archive_sha, "references": [i["reference"] for i in expected]})
        validation.append("PASS %s real archive config IDs and copied SHA256" % role)
    write_json(package / "images/archives.json", {"schemaVersion": 1, "archives": archive_entries})
    if diagnostic_profile:
        write_json(package / "provenance/diagnostics-lock.json", diagnostic_profile)
    # Preserve reviewed history away from runnable user entrypoints.
    if (baseline / "source").is_dir():
        shutil.copytree(baseline / "source", package / "provenance/baseline", ignore=shutil.ignore_patterns("__pycache__", "*.pyc"))
    if (baseline / "evidence").is_dir():
        shutil.copytree(baseline / "evidence", package / "provenance/baseline-evidence")
    shutil.copyfile(Path(__file__).resolve(), package / "provenance/build-package.py")
    source_tar = package / "provenance/current-source.tar"
    if source_export is None:
        run(["git", "archive", "--format=tar", "--output=" + str(source_tar), "HEAD"], cwd=source)
        source_export = source_tar
    with source_export.open("rb") as src:
        is_gzip = src.read(2) == b"\x1f\x8b"
    if is_gzip:
        shutil.copyfile(source_export, package / "provenance/current-source.tar.gz")
    else:
        with source_export.open("rb") as src, (package / "provenance/current-source.tar.gz").open("wb") as dst:
            with gzip.GzipFile(filename="", mode="wb", fileobj=dst, mtime=0, compresslevel=1) as compressed:
                shutil.copyfileobj(src, compressed)
    if source_tar.exists():
        source_tar.unlink()
    versions = patch_charts(source, package, by_role, timestamp)
    quickstart_profile(baseline, package, by_role)
    shutil.copytree(baseline / "prerequisites", package / "prerequisites")
    shutil.copyfile(source / "install.sh", package / "install.sh")
    (package / "install.sh").chmod(0o755)
    (package / "bin").mkdir()
    shutil.copyfile(installer, package / "bin/lakehouse-installer")
    (package / "bin/lakehouse-installer").chmod(0o755)
    shutil.copyfile(source / "LICENSE", package / "LICENSE")
    if "build_host" in lock:
        lock.setdefault("binary_build_host", lock.pop("build_host"))
    lock.update(source_sha=source_sha, chart_source_sha=source_sha, chart_versions=versions,
                build_date=datetime.now().astimezone().strftime("%Y-%m-%d"), build_host=socket.gethostname(),
                scope="internal single-worker QA; base images retained; runtime image identities and diagnostic overrides recorded per role")
    lock.setdefault("packaging", {}).update({"version": args.version, "installer_source_sha": source_sha,
        "installer_sha256": digest(installer), "binary_rebuild": False, "baseline_source_sha": baseline_sha,
        "source_archive": "provenance/current-source.tar.gz",
        "source_archive_sha256": digest(package / "provenance/current-source.tar.gz"),
        "archive_manifest": "images/archives.json", "minio_mc": "included in bundle: images/" + minio_name})
    if diagnostic_profile:
        lock["packaging"].update({"diagnostics_lock": "provenance/diagnostics-lock.json",
                                  "updated_image_roles": sorted(updated_roles),
                                  "binary_rebuild": bool(diagnostic_profile.get("binary_rebuild", False))})
    for image in lock["images"]:
        if image["role"] in MINIO_ROLES:
            image["delivery"] = "archive"
    write_json(package / "image-lock.json", lock)
    instructions = (source / "INSTALL.txt").read_text()
    (package / "INSTALL.txt").write_text(instructions)
    instructions = instructions.replace("oceanbase-lakehouse-具体版本.tar", name + ".tar")
    instructions = instructions.replace("oceanbase-lakehouse-具体版本", name)
    (package / "README.txt").write_text("交付版本：%s\n源码：%s\n所有镜像和校验清单均已包含在本 tar 内。\n\n" %
        (args.version, source_sha) + instructions)
    validation += validate_charts(package, lock)
    run(["bash", "-n", str(package / "install.sh")])
    validation.append("PASS root shell entry syntax (user runtime requires no Python)")
    write_json(package / "provenance/build-validation.json", {"source_sha": source_sha,
               "binary_source_sha": lock.get("binary_source_sha"), "operator_source_sha": lock.get("operator_source_sha"),
               "checks": validation, "live_installation": "not performed by the package builder"})
    checksums = []
    for path in sorted(package.rglob("*")):
        if path.is_file():
            checksums.append(digest(path) + "  " + path.relative_to(package).as_posix())
    (package / "SHA256SUMS").write_text("\n".join(checksums) + "\n")
    full_tar = output / (name + ".tar")
    tar_tree(package, full_tar, name, timestamp)
    full_sha = digest(full_tar)
    minio_sha = next(entry["sha256"] for entry in archive_entries if entry["name"] == "minio")
    (output / "README.txt").write_text("用户交付只提供下面一个文件：\n" + full_tar.name +
        "\n\ntar -xf 解压，进入目录执行 ./install.sh。所有镜像（包括 MinIO/mc）和 SHA256SUMS 均在包内。\n"
        "本目录的构建记录仅供内部追溯，不是安装依赖；不需要外部校验文件或镜像侧包。\n")
    result = {"source_sha": source_sha, "package": str(full_tar), "sha256": full_sha,
              "size_bytes": full_tar.stat().st_size, "minio_cache": str(package / "images" / minio_name),
              "minio_sha256": minio_sha, "validation": validation}
    write_json(output / "build-result.json", result)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, required=True, help="Clean source checkout with reviewed charts and root install.sh")
    parser.add_argument("--baseline", type=Path, required=True, help="Extracted previous QA bundle (contains images and image-lock.json)")
    parser.add_argument("--output", type=Path, required=True, help="New delivery directory; must not already exist")
    parser.add_argument("--minio-cache", type=Path, required=True, help="Build input: locked MinIO/mc Docker-save gzip archive, copied inside the final tar")
    parser.add_argument("--diagnostics-archive", type=Path, help="Docker-save gzip archive with the exact diagnostics-lock images")
    parser.add_argument("--diagnostics-lock", type=Path, help="Fixed identities for obdiag/helper and optional updated Dashboard or fresh-install Observer")
    parser.add_argument("--version", required=True, help="Unique delivery version, e.g. 20261021-COMMITSHA")
    parser.add_argument("--installer-binary", type=Path, help="Linux/amd64 binary; default SOURCE/bin/lakehouse-installer")
    parser.add_argument("--source-archive", type=Path, help="Original git archive for an extracted source tree; verifies its commit and every exported file without requiring .git")
    args = parser.parse_args()
    try:
        build(args)
    except (ValueError, RuntimeError, OSError, KeyError, subprocess.SubprocessError) as error:
        parser.exit(1, "Package build failed: %s\n" % error)


if __name__ == "__main__":
    main()

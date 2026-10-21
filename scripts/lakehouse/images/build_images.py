#!/usr/bin/env python3
"""Build and verify the pinned Linux/amd64 diagnostic images, without a cluster.

Prepare the files from inputs.lock.json in an input directory, then run:
  python3 build_images.py --inputs INPUT_DIRECTORY --output NEW_OUTPUT_DIRECTORY
Docker 26's legacy builder accepts the locally verified config IDs as FROM.
No image is pulled and all build/smoke-test containers have network disabled.
"""
import argparse
import gzip
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile


HERE = Path(__file__).resolve().parent
REFERENCES = {
    "observer": "oceanbase-qa/observer:4.6.2.0-tar1.30-20261021",
    "obdiag": "oceanbase-qa/obdiag:5.0.1-32026101919-20261021",
    "oceanbase-helper": "oceanbase-qa/oceanbase-helper:93d77619-20261021",
}
RUNTIME_FIELDS = ("User", "Env", "Entrypoint", "Cmd", "WorkingDir", "ExposedPorts",
                  "Volumes", "StopSignal", "Healthcheck", "Labels")


def digest(path):
    result = hashlib.sha256()
    with Path(path).open("rb") as stream:
        for data in iter(lambda: stream.read(8 * 1024 * 1024), b""):
            result.update(data)
    return result.hexdigest()


def run(*args, **kwargs):
    try:
        return subprocess.run(args, check=True, stdout=subprocess.PIPE,
                              stderr=subprocess.STDOUT, text=True, timeout=900, **kwargs).stdout
    except subprocess.CalledProcessError as error:
        raise RuntimeError("Offline image command failed:\n" + error.stdout) from error


def inspect(reference):
    value = json.loads(run("docker", "image", "inspect", reference))[0]
    if value["Os"] != "linux" or value["Architecture"] != "amd64":
        raise ValueError("Expected Linux/amd64 image: " + reference)
    return value


def smoke(reference, command, *args):
    return run("docker", "run", "--rm", "--network", "none", "--entrypoint",
               command, reference, *args)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--inputs", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    args.inputs = args.inputs.resolve()
    source = json.loads((HERE / "inputs.lock.json").read_text())
    for name, entry in source["inputs"].items():
        if digest(args.inputs / name) != entry["sha256"]:
            raise ValueError("Input SHA256 mismatch: " + name)
    bases = {role: inspect(source["observer_base" if role == "observer" else "runtime_base"])
             for role in REFERENCES}
    args.output.mkdir(parents=True, exist_ok=False)
    build_env = dict(os.environ, DOCKER_BUILDKIT="0")
    images = []
    for role, reference in REFERENCES.items():
        dockerfile = HERE / ("Dockerfile.helper" if role == "oceanbase-helper" else "Dockerfile." + role)
        command = ["docker", "build", "--network", "none", "--pull=false", "-f", str(dockerfile),
                   "--build-arg", "BASE_IMAGE=" + bases[role]["Id"], "-t", reference]
        if role == "oceanbase-helper":
            command += ["--build-arg", "SOURCE_REVISION=" + source["helper_revision"]]
        with (args.output / (role + "-build.log")).open("w") as log:
            subprocess.run(command + [str(args.inputs)], env=build_env, stdout=log,
                           stderr=subprocess.STDOUT, check=True, timeout=900)
        image = inspect(reference)
        base_layers = bases[role]["RootFS"]["Layers"]
        if image["RootFS"]["Layers"][:len(base_layers)] != base_layers:
            raise ValueError("Derived image lost base layers: " + role)
        if role == "observer":
            for key in RUNTIME_FIELDS:
                if image["Config"].get(key) != bases[role]["Config"].get(key):
                    raise ValueError("Observer runtime configuration changed: " + key)
            binary = smoke(reference, "sha256sum", "/home/admin/oceanbase/bin/observer").split()[0]
            if binary != source["observer_binary_sha256"]:
                raise ValueError("Observer database binary changed")
            smoke_output = smoke(reference, "/bin/sh", "-ec",
                "tar --version; gzip --version; mkdir -p /tmp/logcheck; "
                "printf 'diagnostic archive probe\\n' > /tmp/logcheck/observer.log; "
                "tar -czf /tmp/logcheck.tar.gz -C /tmp logcheck; "
                "tar cf /tmp/transport.tar -C /tmp logcheck.tar.gz; "
                "mkdir /tmp/unpack; tar xf /tmp/transport.tar -C /tmp/unpack; "
                "tar -xzf /tmp/unpack/logcheck.tar.gz -C /tmp/unpack; "
                "test \"$(cat /tmp/logcheck/observer.log)\" = \"$(cat /tmp/unpack/logcheck/observer.log)\"")
            smoke_output += "\nObserver SHA256: " + binary + "\nRuntime configuration preserved.\n"
        elif role == "obdiag":
            smoke_output = smoke(reference, "/bin/sh", "-ec",
                "obdiag --version; obdiag gather log --help; tar --version; gzip --version; "
                "python3 -c 'import sys, json, tarfile, gzip, pathlib, subprocess, tempfile; "
                "assert sys.version_info >= (3, 6); print(sys.version)'")
            for expected in ("OceanBase Diagnostic Tool: 5.0.1", "--from", "--to", "--store_dir", "tar (GNU tar)"):
                if expected not in smoke_output:
                    raise ValueError("Missing obdiag command capability: " + expected)
        else:
            # The helper initializes Kubernetes clients even for --help. Supply
            # a dummy offline config, never a real cluster credential.
            with tempfile.TemporaryDirectory() as temporary:
                kubeconfig = Path(temporary) / "config"
                kubeconfig.write_text("apiVersion: v1\nkind: Config\nclusters:\n- name: offline\n"
                    "  cluster:\n    server: https://127.0.0.1:1\ncontexts:\n- name: offline\n"
                    "  context:\n    cluster: offline\n    user: offline\ncurrent-context: offline\n"
                    "users:\n- name: offline\n  user: {}\n")
                smoke_output = run("docker", "run", "--rm", "--read-only", "--network", "none",
                    "-v", str(kubeconfig) + ":/tmp/offline-kubeconfig:ro", "-e", "KUBECONFIG=/tmp/offline-kubeconfig",
                    "--entrypoint", "/home/admin/oceanbase/bin/oceanbase-helper", reference,
                    "generate", "obdiag-config", "--help")
            for expected in ("--namespace", "--cluster", "--output"):
                if expected not in smoke_output:
                    raise ValueError("Missing helper config command option: " + expected)
            binary = smoke(reference, "sha256sum", "/home/admin/oceanbase/bin/oceanbase-helper").split()[0]
            if binary != source["inputs"]["oceanbase-helper"]["sha256"]:
                raise ValueError("Helper binary changed")
        (args.output / (role + "-smoke.log")).write_text(smoke_output)
        images.append({"role": role, "reference": reference, "image_id": image["Id"],
                       "base_image_id": bases[role]["Id"], "source": source})
    archive = args.output / "diagnostics-base-images.tar.gz"
    temporary_tar = args.output / "diagnostics-base-images.tar"
    run("docker", "image", "save", "--output", str(temporary_tar), *REFERENCES.values())
    with temporary_tar.open("rb") as original, archive.open("xb") as raw:
        with gzip.GzipFile(fileobj=raw, mode="wb", filename="", mtime=0, compresslevel=1) as compressed:
            shutil.copyfileobj(original, compressed, 8 * 1024 * 1024)
    temporary_tar.unlink()
    manifest = {"schemaVersion": 1, "archive_sha256": digest(archive), "images": images}
    (args.output / "diagnostics-lock.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(json.dumps({"archive": str(archive), "archive_sha256": manifest["archive_sha256"],
                      "images": [{key: image[key] for key in ("role", "reference", "image_id")} for image in images]}, indent=2))


if __name__ == "__main__":
    main()

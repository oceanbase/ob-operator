"""Small local packaging regressions; no cluster or real image cache is used.

Run with: python3 -m unittest discover -s scripts/lakehouse -p 'test_*.py'
The full-build fixture requires the documented build-host Helm/Git dependencies.
"""
import argparse
import contextlib
import copy
import gzip
import hashlib
import importlib.util
import io
import json
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile
import unittest

import yaml


HERE = Path(__file__).resolve().parent
REPO = HERE.parent.parent
SPEC = importlib.util.spec_from_file_location("lakehouse_package_builder", HERE / "build-package.py")
BUILDER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(BUILDER)

REFERENCES = {
    "operator": "oceanbase-qa/ob-operator:96f9d3c8-20261019",
    "dashboard": "oceanbase-qa/oceanbase-dashboard:e747a05a-20261019",
    "observer": "oceanbase-qa/observer:4.6.2.0-e747a05a-20261019",
    "logservice": "oceanbase-qa/oblogservice:1.3.0-e747a05a-20261019",
    "prometheus": "quay.io/prometheus/prometheus:v2.51.1",
    "alertmanager": "quay.io/prometheus/alertmanager:v0.27.0",
    "busybox": "busybox:latest",
    "obagent": "oceanbase/obagent:4.2.4-200000022025090416",
    "cert-manager-controller": "quay.io/jetstack/cert-manager-controller:v1.12.4",
    "cert-manager-cainjector": "quay.io/jetstack/cert-manager-cainjector:v1.12.4",
    "cert-manager-webhook": "quay.io/jetstack/cert-manager-webhook:v1.12.4",
    "minio": "minio/minio:RELEASE.2025-09-07T16-13-09Z",
    "mc": "minio/mc:RELEASE.2025-08-13T08-35-41Z",
    "kube-rbac-proxy": "oceanbase/kube-rbac-proxy:v0.13.1",
}


def add_bytes(archive, name, payload):
    entry = tarfile.TarInfo(name)
    entry.size = len(payload)
    archive.addfile(entry, io.BytesIO(payload))


def make_images():
    images, configs = [], {}
    for role, reference in REFERENCES.items():
        payload = json.dumps({"os": "linux", "architecture": "amd64",
                              "config": {"Labels": {"fixture": role}},
                              "rootfs": {"type": "layers", "diff_ids": []}}, sort_keys=True).encode()
        images.append({"role": role, "reference": reference,
                       "image_id": "sha256:" + hashlib.sha256(payload).hexdigest()})
        configs[reference] = payload
    return images, configs


def docker_save(path, images, configs, oci=True):
    with path.open("wb") as raw, gzip.GzipFile(filename="", fileobj=raw, mode="wb", mtime=0) as compressed:
        with tarfile.open(fileobj=compressed, mode="w|", format=tarfile.GNU_FORMAT) as archive:
            manifest = []
            for image in images:
                payload = configs[image["reference"]]
                sha = hashlib.sha256(payload).hexdigest()
                name = "blobs/sha256/" + sha if oci else sha + ".json"
                add_bytes(archive, name, payload)
                manifest.append({"Config": name, "RepoTags": [image["reference"]], "Layers": []})
            add_bytes(archive, "manifest.json", json.dumps(manifest).encode())


class BuildPackageTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="lakehouse-package-test-")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.images, self.configs = make_images()

    def test_docker_save_config_ids_and_integrity(self):
        for oci in (True, False):
            with self.subTest(oci=oci):
                archive = self.root / ("oci.tar.gz" if oci else "legacy.tar.gz")
                docker_save(archive, self.images, self.configs, oci)
                expected = {image["reference"]: image["image_id"] for image in self.images}
                self.assertEqual(BUILDER.archive_image_ids(archive), expected)
                self.assertEqual(BUILDER.verify_archive(archive, self.images), BUILDER.digest(archive))
                invalid = copy.deepcopy(self.images)
                invalid[0]["image_id"] = "sha256:" + "0" * 64
                with self.assertRaisesRegex(ValueError, "image ID mismatch"):
                    BUILDER.verify_archive(archive, invalid)
                with self.assertRaisesRegex(ValueError, "SHA256 mismatch"):
                    BUILDER.verify_archive(archive, self.images, "0" * 64)

    @unittest.skipUnless(shutil.which("git"), "Git is a build-host dependency")
    def test_source_archive_verification(self):
        source = self.make_source(charts=False)
        exported = self.root / "source.tar"
        BUILDER.run(["git", "archive", "--format=tar", "--output=" + str(exported), "HEAD"], cwd=source)
        revision, epoch = BUILDER.source_identity(source, exported)
        self.assertEqual(revision, BUILDER.run(["git", "rev-parse", "HEAD"], cwd=source).strip())
        self.assertGreater(epoch, 0)
        (source / "INSTALL.txt").write_text("modified after export")
        with self.assertRaisesRegex(ValueError, "differs from git archive"):
            BUILDER.source_identity(source, exported)
        malicious = self.root / "unsafe.tar"
        with tarfile.open(malicious, "w", format=tarfile.PAX_FORMAT,
                          pax_headers={"comment": revision}) as archive:
            add_bytes(archive, "../outside", b"must never be extracted")
        with self.assertRaisesRegex(ValueError, "Unsafe source archive path"):
            BUILDER.source_identity(source, malicious)
        with tarfile.open(self.root / "no-identity.tar", "w") as archive:
            add_bytes(archive, "INSTALL.txt", b"no git identity")
        with self.assertRaisesRegex(ValueError, "original commit PAX comment"):
            BUILDER.source_identity(source, self.root / "no-identity.tar")

    def make_source(self, charts=True):
        source = self.root / "source"
        source.mkdir()
        if charts:
            for name in ("oceanbase-lakehouse", "oceanbase-dashboard", "ob-operator"):
                shutil.copytree(REPO / "charts" / name, source / "charts" / name)
        for name in ("install.sh", "INSTALL.txt", "LICENSE"):
            shutil.copyfile(REPO / name, source / name)
        BUILDER.run(["git", "init", "-q"], cwd=source)
        BUILDER.run(["git", "add", "."], cwd=source)
        BUILDER.run(["git", "-c", "user.name=Package Fixture", "-c", "user.email=fixture@example.invalid",
                     "-c", "commit.gpgsign=false", "commit", "-qm", "fixture source"], cwd=source)
        return source

    @unittest.skipUnless(shutil.which("git") and shutil.which("helm"), "Full build requires Helm and Git")
    def test_complete_delivery_export_and_reproducibility(self):
        source = self.make_source()
        baseline = self.root / "baseline"
        for directory in ("images", "values", "prerequisites"):
            (baseline / directory).mkdir(parents=True)
        lock = {"source_sha": "a" * 40, "binary_source_sha": "b" * 40,
                "operator_source_sha": "c" * 40, "build_host": "original-binary-host",
                "packaging": {"source_archive": "old/source.tar.gz", "source_archive_sha256": "0" * 64},
                "images": self.images}
        BUILDER.write_json(baseline / "image-lock.json", lock)
        values = yaml.safe_load((source / "charts/oceanbase-lakehouse/values.yaml").read_text())
        BUILDER.write_yaml(baseline / "values/single-worker.yaml", values)
        cert = {"apiVersion": "v1", "kind": "Pod", "metadata": {"name": "fixture"},
                "spec": {"containers": [{"name": role, "image": reference} for role, reference in REFERENCES.items()
                                         if role.startswith("cert-manager-")]}}
        BUILDER.write_yaml(baseline / "prerequisites/cert-manager-v1.12.4.yaml", cert)
        for name, roles in (("core", BUILDER.CORE_ROLES),
                            ("support", REFERENCES.keys() - BUILDER.CORE_ROLES - BUILDER.MINIO_ROLES)):
            docker_save(baseline / "images" / (name + "-images.tar.gz"),
                        [image for image in self.images if image["role"] in roles], self.configs)
        (baseline / "SHA256SUMS").write_text("".join(BUILDER.digest(path) + "  " +
            path.relative_to(baseline).as_posix() + "\n" for path in sorted(baseline.rglob("*")) if path.is_file()))
        minio_cache = self.root / "fixture-minio.tar.gz"
        docker_save(minio_cache, [image for image in self.images if image["role"] in BUILDER.MINIO_ROLES], self.configs)
        binary = self.root / "fixture-installer"
        binary.write_bytes(b"\x7fELF\x02\x01" + b"\0" * 12 + b"\x3e\0")  # fixture architecture header only
        args = argparse.Namespace(source=source, baseline=baseline, output=self.root / "delivery-a",
                                  minio_cache=minio_cache, version="fixture", installer_binary=binary, source_archive=None)
        with contextlib.redirect_stdout(io.StringIO()):
            first = BUILDER.build(args)
        # The remote builder receives an exported source tree, not a Git checkout.
        source_archive = self.root / "source.tar"
        BUILDER.run(["git", "archive", "--format=tar", "--output=" + str(source_archive), "HEAD"], cwd=source)
        exported = self.root / "exported"
        shutil.copytree(source, exported, ignore=shutil.ignore_patterns(".git"))
        args.source, args.source_archive, args.output = exported, source_archive, self.root / "delivery-b"
        with contextlib.redirect_stdout(io.StringIO()):
            second = BUILDER.build(args)
        self.assertEqual(first["sha256"], second["sha256"])
        bundle = args.output / "oceanbase-lakehouse-qa-fixture-linux-amd64"
        manifest = json.loads((bundle / "images/archives.json").read_text())
        sidecar = next(entry for entry in manifest["archives"] if entry["name"] == "minio")
        self.assertEqual(sidecar["base"], "delivery")
        self.assertEqual(sidecar["path"], "minio-cache-fixture-linux-amd64.tar.gz")
        self.assertEqual(BUILDER.digest(bundle.parent / sidecar["path"]), sidecar["sha256"])
        quickstart = yaml.safe_load((bundle / "values/quickstart.yaml").read_text())
        self.assertEqual(quickstart["storageClass"], "")
        self.assertEqual(quickstart["logService"]["resource"]["cpu"], "1")
        self.assertEqual(quickstart["observer"]["resource"]["cpu"], "2")
        self.assertEqual(quickstart["dashboard"]["service"]["type"], "NodePort")
        self.assertNotIn("nodePort", quickstart["dashboard"]["service"])
        for section in ("observer", "logService"):
            self.assertNotIn("clusterId", quickstart[section])
        final_lock = json.loads((bundle / "image-lock.json").read_text())
        for field in ("binary_source_sha", "operator_source_sha"):
            self.assertEqual(final_lock[field], lock[field])
        self.assertEqual(final_lock["binary_build_host"], "original-binary-host")
        self.assertEqual([image["image_id"] for image in final_lock["images"]],
                         [image["image_id"] for image in self.images])
        self.assertEqual(final_lock["source_sha"], first["source_sha"])
        self.assertEqual(final_lock["packaging"]["source_archive"], "provenance/current-source.tar.gz")
        self.assertEqual(BUILDER.digest(bundle / final_lock["packaging"]["source_archive"]),
                         final_lock["packaging"]["source_archive_sha256"])
        self.assertEqual((bundle / "INSTALL.txt").read_bytes(), (source / "INSTALL.txt").read_bytes())
        self.assertNotIn("具体版本", (bundle / "README.txt").read_text())
        self.assertIn("PASS standalone Dashboard password guard", second["validation"])
        self.assertEqual(sum("PASS render stage=" in entry for entry in second["validation"]), 6)
        for line in (bundle / "SHA256SUMS").read_text().splitlines():
            expected, name = line.split("  ", 1)
            self.assertEqual(BUILDER.digest(bundle / name), expected, name)
        # Extra chart files must not silently enter a build attributed to a commit.
        (exported / "charts/oceanbase-lakehouse/untracked.txt").write_text("unexpected input")
        with self.assertRaisesRegex(ValueError, "Untracked input"):
            BUILDER.source_identity(exported, source_archive)


if __name__ == "__main__":
    unittest.main()

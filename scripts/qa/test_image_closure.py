import gzip
import hashlib
import importlib.util
import io
import json
from pathlib import Path
import tarfile
import tempfile
import unittest

spec = importlib.util.spec_from_file_location("closure", Path(__file__).with_name("verify-image-closure.py"))
closure = importlib.util.module_from_spec(spec)
spec.loader.exec_module(closure)


def entry(ref, role="operator", delivery="archive", image_id=None):
    return dict(reference=ref, role=role, delivery=delivery, image_id=image_id or "sha256:" + "a" * 64)


def saved_image(path, ref, oci=True, false_config_name=False):
    config = json.dumps({"os": "linux", "architecture": "amd64", "config": {"fixture": ref}}).encode()
    sha = hashlib.sha256(config).hexdigest()
    filename = "0" * 64 if false_config_name else sha
    name = "blobs/sha256/" + filename if oci else filename + ".json"
    manifest = json.dumps([{"Config": name, "RepoTags": [ref], "Layers": []}]).encode()
    with tarfile.open(path, "w:gz") as archive:
        for member, payload in ((name, config), ("manifest.json", manifest)):
            info = tarfile.TarInfo(member)
            info.size = len(payload)
            archive.addfile(info, io.BytesIO(payload))
    return "sha256:" + sha


class ImageClosureTest(unittest.TestCase):
    def test_missing_sidecar_is_a_hard_failure(self):
        manifest = "spec:\n  containers:\n  - image: operator:qa\n  - image: oceanbase/kube-rbac-proxy:v0.13.1\n"
        with self.assertRaisesRegex(ValueError, "kube-rbac-proxy"):
            closure.verify({"images": [entry("operator:qa")]}, [manifest])
        self.assertEqual(closure.verify({"images": [entry("operator:qa"), entry("oceanbase/kube-rbac-proxy:v0.13.1", "kube-rbac-proxy")]}, [manifest]), 2)

    def test_nested_and_multidoc_manifests(self):
        self.assertEqual(closure.verify({"images": [entry("operator:qa")]}, ["spec: {template: {spec: {initContainers: [{image: 'docker.io/library/operator:qa'}]}}}\n---\nspec: {logService: {image: 'operator:qa'}}"]), 1)

    def test_minio_delivery_policy_and_empty_inputs(self):
        with self.assertRaisesRegex(ValueError, "outside the main archive"):
            closure.verify({"images": [entry("minio:qa", "minio")]}, ["image: minio:qa"])
        with self.assertRaisesRegex(ValueError, "No rendered"):
            closure.verify({"images": []}, ["kind: ConfigMap"])

    def test_archive_content_and_config_bytes_are_checked(self):
        with tempfile.TemporaryDirectory() as tmp:
            for oci in (True, False):
                with self.subTest(oci=oci):
                    path = Path(tmp) / "images.tar.gz"
                    image_id = saved_image(path, "operator:qa", oci=oci, false_config_name=True)
                    locked = entry("operator:qa", image_id=image_id)
                    self.assertEqual(closure.verify({"images": [locked]}, ["image: operator:qa"], [path]), 1)
                    with self.assertRaisesRegex(ValueError, "Missing/mismatched"):
                        closure.verify({"images": [entry("operator:qa", image_id="sha256:" + "0" * 64)]}, ["image: operator:qa"], [path])
                    with self.assertRaisesRegex(ValueError, "Missing/mismatched"):
                        closure.verify({"images": [locked, entry("proxy:qa")]}, ["image: operator:qa"], [path])
                    locked["delivery"] = "online-or-matching-existing-cache"
                    with self.assertRaisesRegex(ValueError, "Online-only"):
                        closure.verify({"images": [locked]}, ["image: operator:qa"], [path])

    def test_configmap_diagnostic_jobs_and_separate_minio_cache(self):
        manifest = "kind: ConfigMap\ndata:\n  config.yaml: |\n    inspection:\n      obdiag: {image: 'obdiag:fixed'}\n      oceanbase-helper: {image: 'helper:fixed'}\n    sqlAnalyzer: {image: 'optional/analyzer:1'}\n"
        with self.assertRaisesRegex(ValueError, "helper:fixed|obdiag:fixed"):
            closure.verify({"images": [entry("operator:qa")]}, [manifest])
        with tempfile.TemporaryDirectory() as tmp:
            paths, images = [], []
            for ref, role in (("obdiag:fixed", "obdiag"), ("helper:fixed", "oceanbase-helper")):
                path = Path(tmp) / (role + ".tar.gz")
                images.append(entry(ref, role, image_id=saved_image(path, ref)))
                paths.append(path)
            minio = Path(tmp) / "minio.tar.gz"
            images.append(entry("minio:fixed", "minio", "delivery-sidecar", saved_image(minio, "minio:fixed")))
            self.assertEqual(closure.verify({"images": images}, [manifest], paths, [minio]), 2)
            with self.assertRaisesRegex(ValueError, "delivery-sidecar"):
                closure.verify({"images": images}, [manifest], paths)
            with self.assertRaisesRegex(ValueError, "wrong delivery archive|Missing/mismatched"):
                closure.verify({"images": images}, [manifest], paths + [minio], [minio])
            with self.assertRaisesRegex(ValueError, "Untracked archived"):
                closure.verify({"images": images[:-1]}, [manifest], paths + [minio])

    def test_manifest_only_archive_is_not_image_identity_evidence(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "fake.tar.gz"
            data = json.dumps([{"Config": "a" * 64 + ".json", "RepoTags": ["operator:qa"]}]).encode()
            with tarfile.open(path, "w:gz") as archive:
                info = tarfile.TarInfo("manifest.json")
                info.size = len(data)
                archive.addfile(info, io.BytesIO(data))
            with self.assertRaisesRegex(ValueError, "Missing image config"):
                closure.verify({"images": [entry("operator:qa")]}, ["image: operator:qa"], [path])

    def test_uncompressed_docker_save_is_still_supported(self):
        with tempfile.TemporaryDirectory() as tmp:
            compressed, plain = Path(tmp) / "images.tar.gz", Path(tmp) / "images.tar"
            image_id = saved_image(compressed, "operator:qa")
            with gzip.open(compressed, "rb") as stream:
                plain.write_bytes(stream.read())
            self.assertEqual(closure.verify({"images": [entry("operator:qa", image_id=image_id)]},
                                            ["image: operator:qa"], [plain]), 1)


if __name__ == "__main__":
    unittest.main()

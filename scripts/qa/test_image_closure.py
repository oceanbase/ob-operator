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


def entry(ref, role="operator", delivery="archive"):
    return dict(reference=ref, role=role, delivery=delivery, image_id="sha256:" + "a" * 64)


class ImageClosureTest(unittest.TestCase):
    def test_missing_sidecar_is_a_hard_failure(self):
        manifest = "spec:\n  containers:\n  - image: operator:qa\n  - image: oceanbase/kube-rbac-proxy:v0.13.1\n"
        with self.assertRaisesRegex(ValueError, "kube-rbac-proxy"):
            closure.verify({"images": [entry("operator:qa")]}, [manifest])
        self.assertEqual(closure.verify({"images": [entry("operator:qa"), entry("oceanbase/kube-rbac-proxy:v0.13.1", "kube-rbac-proxy")]}, [manifest]), 2)

    def test_nested_and_multidoc_manifests(self):
        self.assertEqual(closure.verify({"images": [entry("operator:qa")]}, ["spec: {template: {spec: {initContainers: [{image: 'operator:qa'}]}}}\n---\nspec: {logService: {image: 'operator:qa'}}"]), 1)

    def test_online_policy_and_empty_inputs(self):
        with self.assertRaisesRegex(ValueError, "online-only"):
            closure.verify({"images": [entry("minio:qa", "minio")]}, ["image: minio:qa"])
        with self.assertRaisesRegex(ValueError, "No rendered"):
            closure.verify({"images": []}, ["kind: ConfigMap"])

    def test_archive_content_and_ids_are_checked(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "images.tar.gz"
            data = json.dumps([{"Config": "a"*64+".json", "RepoTags": ["operator:qa"]}]).encode()
            with tarfile.open(path, "w:gz") as archive:
                info = tarfile.TarInfo("manifest.json")
                info.size = len(data)
                archive.addfile(info, io.BytesIO(data))
            self.assertEqual(closure.verify({"images": [entry("operator:qa")]}, ["image: operator:qa"], [path]), 1)
            with self.assertRaisesRegex(ValueError, "Missing/mismatched"):
                closure.verify({"images": [entry("operator:qa"), entry("proxy:qa")]}, ["image: operator:qa"], [path])
            with self.assertRaisesRegex(ValueError, "Online-only"):
                closure.verify({"images": [entry("operator:qa", delivery="online-or-matching-existing-cache")]}, ["image: operator:qa"], [path])


if __name__ == "__main__":
    unittest.main()

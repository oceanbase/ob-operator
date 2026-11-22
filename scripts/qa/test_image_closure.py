import gzip
import hashlib
import importlib.util
import io
import json
from pathlib import Path
import shutil
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
        self.assertEqual(closure.verify({"images": [entry("minio:qa", "minio")]}, ["image: minio:qa"]), 1)
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

    def test_empty_configmap_data_does_not_hide_other_images(self):
        baseline = Path(__file__).with_name("fixtures") / "cert-manager-v1.12.4-empty-configmap.yaml"
        self.assertEqual(list(closure.image_refs(closure.yaml.safe_load(baseline.read_text()))), [])
        for data in ("", "data: null", "data: {}"):
            with self.subTest(data=data):
                manifest = "kind: ConfigMap\n" + data + "\n---\nimage: operator:qa\n"
                self.assertEqual(closure.verify({"images": [entry("operator:qa")]}, [manifest]), 1)

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

    def test_single_tar_bundle_checks_hashes_ids_and_contained_paths(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = root / "built" / "bundle"
            (source / "images").mkdir(parents=True)
            images, archives = [], []
            for role in ("operator", "minio", "mc"):
                path = source / "images" / (role + ".tar.gz")
                ref = role + ":qa"
                images.append(entry(ref, role, image_id=saved_image(path, ref)))
                archives.append({"name": role, "base": "bundle", "path": "images/" + path.name,
                                 "sha256": hashlib.sha256(path.read_bytes()).hexdigest(), "references": [ref]})
            (source / "image-lock.json").write_text(json.dumps({"images": images}))
            (source / "images/archives.json").write_text(json.dumps({"schemaVersion": 1, "archives": archives}))
            def checksums(bundle):
                (bundle / "SHA256SUMS").write_text("".join(hashlib.sha256(path.read_bytes()).hexdigest() + "  " +
                    path.relative_to(bundle).as_posix() + "\n" for path in sorted(bundle.rglob("*"))
                    if path.is_file() and path != bundle / "SHA256SUMS"))
            checksums(source)
            package = root / "bundle.tar"
            with tarfile.open(package, "w") as archive:
                archive.add(source, arcname="bundle")
            recipient = root / "recipient"
            recipient.mkdir()
            shutil.copyfile(package, recipient / package.name)
            self.assertEqual([path.name for path in recipient.iterdir()], ["bundle.tar"])
            with tarfile.open(recipient / package.name) as archive:
                archive.extractall(recipient)
            delivered = recipient / "bundle"
            manifest = "image: operator:qa\n---\nimage: minio:qa\n---\nimage: mc:qa"
            self.assertEqual(closure.verify_bundle(delivered, [manifest]), 3)
            # Mutations below are applied independently to the real extracted
            # artifact. Metadata checksums are updated when testing deeper
            # identity/path contracts rather than just corruption detection.
            for case in ("missing", "corrupt", "wrong-id", "wrong-sha", "escape", "sidecar", "refs", "symlink"):
                with self.subTest(case=case):
                    changed = root / case
                    shutil.copytree(delivered, changed)
                    minio = changed / "images/minio.tar.gz"
                    metadata = json.loads((changed / "images/archives.json").read_text())
                    if case == "missing":
                        minio.unlink()
                    elif case == "corrupt":
                        minio.write_bytes(minio.read_bytes() + b"corrupt")
                    elif case == "wrong-id":
                        locked = json.loads((changed / "image-lock.json").read_text())
                        locked["images"][1]["image_id"] = "sha256:" + "0" * 64
                        (changed / "image-lock.json").write_text(json.dumps(locked))
                    elif case == "wrong-sha":
                        metadata["archives"][1]["sha256"] = "0" * 64
                    elif case == "escape":
                        metadata["archives"][1]["path"] = "../outside.tar.gz"
                        shutil.copyfile(minio, root / "outside.tar.gz")
                    elif case == "sidecar":
                        metadata["archives"][1]["base"] = "delivery"
                    elif case == "refs":
                        metadata["archives"][1]["references"] = ["mc:qa"]
                    elif case == "symlink":
                        outside = root / "linked-image.tar.gz"
                        shutil.copyfile(minio, outside)
                        minio.unlink()
                        minio.symlink_to(outside)
                    if case not in ("missing", "corrupt"):
                        (changed / "images/archives.json").write_text(json.dumps(metadata))
                        checksums(changed)
                    with self.assertRaises((ValueError, OSError)):
                        closure.verify_bundle(changed, [manifest])


if __name__ == "__main__":
    unittest.main()

"""Chart contract tests. Requires Helm 3, PyYAML, and jsonschema."""

import copy
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

import jsonschema
import yaml


CHART = Path(__file__).resolve().parents[1]
ROOT = CHART.parents[1]
HELM = os.environ.get("HELM", "helm")
LAKEHOUSE = {
    "enabled": True,
    "observerImage": "example.invalid/observer:test",
    "logService": {
        "image": "example.invalid/logservice:test",
        "bucketURL": "s3://logs?host=http://minio:9000&s3_region=us-east-1",
    },
    "sharedStorage": {
        "bucketURL": "s3://data?host=http://minio:9000&s3_region=us-east-1",
    },
    "credentials": {"accessId": "test-id", "accessKey": "test-key"},
}


class RenderTests(unittest.TestCase):
    def render(self, values=None, error=None):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "values.yaml"
            path.write_text(yaml.safe_dump(values or {}))
            result = subprocess.run(
                [HELM, "template", "lakehouse-test", str(CHART),
                 "--namespace", "chart-test", "--values", str(path)],
                capture_output=True, text=True, check=False,
            )
        if error:
            self.assertNotEqual(result.returncode, 0)
            self.assertIn(error, result.stderr)
            return []
        self.assertEqual(result.returncode, 0, result.stderr)
        resources = [item for item in yaml.safe_load_all(result.stdout) if item]
        # Validate the rendered spec against this repository's actual CRDs.
        for item in resources:
            if item["kind"] not in ("OBCluster", "OBLogServiceCluster"):
                continue
            plural = item["kind"].lower() + "s"
            crd = yaml.safe_load((ROOT / "config/crd/bases" /
                                  f"oceanbase.oceanbase.com_{plural}.yaml").read_text())
            version = next(v for v in crd["spec"]["versions"]
                           if v["name"] == "v1alpha1")
            schema = version["schema"]["openAPIV3Schema"]["properties"]["spec"]
            jsonschema.Draft7Validator(schema).validate(item["spec"])
            self.assertEqual(item["metadata"]["namespace"], "chart-test")
        return resources

    @staticmethod
    def resource(items, kind):
        return next(item for item in items if item["kind"] == kind)

    @staticmethod
    def s3_secrets(items):
        return {item["metadata"]["name"]: item for item in items
                if item["kind"] == "Secret" and "stringData" in item}

    def test_traditional_defaults(self):
        items = self.render()
        spec = self.resource(items, "OBCluster")["spec"]
        self.assertNotIn("deploymentMode", spec)
        self.assertIn("redoLogStorage", spec["observer"]["storage"])
        self.assertFalse(any(item["kind"] == "OBLogServiceCluster" for item in items))
        self.assertFalse(self.s3_secrets(items))

    def test_traditional_image_override(self):
        spec = self.resource(self.render({"observerImage": "observer:custom"}), "OBCluster")["spec"]
        self.assertEqual(spec["observer"]["image"], "observer:custom")

    def test_generated_secrets_and_cross_resource_references(self):
        items = self.render({"lakehouse": LAKEHOUSE})
        spec = self.resource(items, "OBCluster")["spec"]
        logservice = self.resource(items, "OBLogServiceCluster")
        self.assertEqual(spec["deploymentMode"], "shared_storage")
        self.assertEqual(spec["observer"]["image"], LAKEHOUSE["observerImage"])
        self.assertEqual(logservice["spec"]["topology"], [
            {"zone": "zone1", "replica": 1},
            {"zone": "zone2", "replica": 1},
            {"zone": "zone3", "replica": 1},
        ])
        self.assertNotIn("redoLogStorage", spec["observer"]["storage"])
        self.assertEqual(spec["logServiceRef"]["name"], logservice["metadata"]["name"])
        secrets = self.s3_secrets(items)
        self.assertEqual(len(secrets), 2)
        for config in (spec["sharedStorageInfo"], logservice["spec"]["objectStoreUrl"]):
            secret = secrets[config["secretRef"]["name"]]
            self.assertEqual(secret["stringData"], {"access_id": "test-id", "access_key": "test-key"})

    def test_existing_and_mixed_secrets(self):
        for log_existing, shared_existing in ((True, True), (True, False), (False, True)):
            with self.subTest(log_existing=log_existing, shared_existing=shared_existing):
                config = copy.deepcopy(LAKEHOUSE)
                if log_existing:
                    config["logService"]["existingSecretName"] = "existing-logs"
                if shared_existing:
                    config["sharedStorage"]["existingSecretName"] = "existing-data"
                if log_existing and shared_existing:
                    config["credentials"] = {"accessId": "", "accessKey": ""}
                items = self.render({"lakehouse": config})
                self.assertEqual(len(self.s3_secrets(items)), 2 - log_existing - shared_existing)
                if log_existing:
                    log = self.resource(items, "OBLogServiceCluster")["spec"]
                    self.assertEqual(log["objectStoreUrl"]["secretRef"]["name"], "existing-logs")
                if shared_existing:
                    ob = self.resource(items, "OBCluster")["spec"]
                    self.assertEqual(ob["sharedStorageInfo"]["secretRef"]["name"], "existing-data")

    def test_custom_logservice_and_io_limits(self):
        config = copy.deepcopy(LAKEHOUSE)
        config["logService"]["nameOverride"] = "custom-logs"
        config["sharedStorage"].update(maxIOPS="1000", maxBandwidth="10M")
        items = self.render({"lakehouse": config, "observerImage": "ignored:normal"})
        spec = self.resource(items, "OBCluster")["spec"]
        self.assertEqual(spec["logServiceRef"]["name"], "custom-logs")
        self.assertEqual(spec["sharedStorageInfo"]["maxIOPS"], "1000")
        self.assertEqual(spec["sharedStorageInfo"]["maxBandwidth"], "10M")
        self.assertEqual(spec["observer"]["image"], config["observerImage"])

    def test_required_inputs(self):
        for section, field in ((None, "observerImage"), ("logService", "image"),
                               ("logService", "bucketURL"), ("sharedStorage", "bucketURL"),
                               ("credentials", "accessId"), ("credentials", "accessKey")):
            with self.subTest(section=section, field=field):
                config = copy.deepcopy(LAKEHOUSE)
                (config[section] if section else config)[field] = ""
                path = ".".join(part for part in ("lakehouse", section, field) if part)
                self.render({"lakehouse": config}, error=path + " is required")

    def test_invalid_bucket_urls(self):
        for section in ("logService", "sharedStorage"):
            for url, error in (
                ("https://bucket?host=http://minio&s3_region=us-east-1", "must start with s3://"),
                ("s3://bucket?s3_region=us-east-1", "non-empty host"),
                ("s3://bucket?host=http://minio", "non-empty s3_region"),
            ):
                with self.subTest(section=section, url=url):
                    config = copy.deepcopy(LAKEHOUSE)
                    config[section]["bucketURL"] = url
                    self.render({"lakehouse": config}, error=error)

    def test_shared_bucket_is_rejected(self):
        config = copy.deepcopy(LAKEHOUSE)
        config["sharedStorage"]["bucketURL"] = "s3://logs?host=http://other&s3_region=other"
        self.render({"lakehouse": config}, error="must use different buckets")


if __name__ == "__main__":
    unittest.main(verbosity=2)

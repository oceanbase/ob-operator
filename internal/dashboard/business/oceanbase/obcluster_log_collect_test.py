"""Fixtures follow obdiag v5.0.1 gather_component_log.py and gather_log/base.py.

The PrettyTable title, columns, node identity, status and payload info are the
upstream format, including the zero-exit failed-node case. No cluster is used.
"""
import importlib.util
import gzip
import io
import json
import os
import tarfile
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location("collector", os.path.join(os.path.dirname(__file__), "obcluster_log_collect.py"))
collector = importlib.util.module_from_spec(spec)
spec.loader.exec_module(collector)


class CollectorTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="log ' path;$ ")
        self.addCleanup(self.temp.cleanup)
        self.work = os.path.join(self.temp.name, "work")
        os.mkdir(self.work)
        self.attachment = os.path.join(self.temp.name, "attachment.tar.gz")
        self.node = "kubernetes_test_ob-pod"
        with open(os.path.join(self.work, "nodes.json"), "w") as stream:
            json.dump([{"namespace": "test", "pod_name": "ob-pod"}], stream)
        with open(os.path.join(self.work, "config.yaml"), "w") as stream:
            stream.write("password: test-secret-must-not-publish")
        os.mkdir(os.path.join(self.work, "k8s"))
        with open(os.path.join(self.work, "k8s", "private.yaml"), "w") as stream:
            stream.write("private-kubeconfig")

    def result(self, rows=None, content=b"[2026-10-21 15:01:00] INFO observer fixture\n", filename="observer.log", corrupt=False):
        pack = os.path.join(self.work, "obdiag_gather_pack_20261021150300")
        os.makedirs(pack, exist_ok=True)
        payload = os.path.join(pack, "observer_log_" + self.node + "_20261021150000_20261021150300_a12345.tar.gz")
        with tarfile.open(payload, "w:gz") as archive:
            entry = tarfile.TarInfo("observer_log_fixture/" + filename)
            entry.size = len(content)
            archive.addfile(entry, io.BytesIO(content))
        if corrupt:
            with open(payload, "wb") as stream:
                stream.write(b"not a tar")
        if rows is None:
            rows = [[self.node, "Success", "12.0 KB", "file save in " + payload]]
        with open(os.path.join(pack, "result_summary.txt"), "w") as stream:
            stream.write("+-------------------------+\n| Gather observer Log Summary on 2026-10-21 15:03:00 |\n+-------------------------+\n")
            stream.write("| Node | Status | Size | info |\n+------+--------+------+------ +\n".replace("------ +", "------+"))
            for row in rows:
                stream.write("| " + " | ".join(row) + " |\n")
            stream.write("+-------------------------+\n")
        return payload

    def execute(self, producer, replace=None):
        def command(argv, **kwargs):
            self.assertEqual(argv, ["obdiag", "gather", "log", "--from", "2026-10-21 15:00:00", "--to", "2026-10-21 15:03:00", "--store_dir", self.work, "-c", os.path.join(self.work, "config.yaml")])
            producer()
            return 0
        with mock.patch.object(collector.subprocess, "call", side_effect=command):
            if replace:
                with mock.patch.object(collector.os, "replace", side_effect=replace):
                    collector.collect("2026-10-21 15:00:00", "2026-10-21 15:03:00", self.work, self.attachment)
            else:
                collector.collect("2026-10-21 15:00:00", "2026-10-21 15:03:00", self.work, self.attachment)

    def assert_rejected(self, producer, exception=collector.CollectionError):
        with self.assertRaises(exception):
            self.execute(producer)
        self.assertFalse(os.path.exists(self.attachment))
        self.assertFalse(os.path.exists(self.work))

    def test_success_quoted_time_special_paths_and_private_publication(self):
        self.execute(self.result)
        self.assertEqual(os.stat(self.attachment).st_mode & 0o777, 0o600)
        self.assertFalse(os.path.exists(self.work))
        with tarfile.open(self.attachment, "r:gz") as archive:
            names = archive.getnames()
            self.assertEqual(len(names), 2)
            self.assertTrue(any(name.endswith("result_summary.txt") for name in names))
            self.assertFalse(any("config" in name or "nodes.json" in name or "output.log" in name for name in names))

    def test_zero_exit_failed_summary_is_failure(self):
        self.assert_rejected(lambda: self.result([[self.node, "Fail", "0", "no log found"]]))

    def test_empty_payload_is_failure(self):
        self.assert_rejected(lambda: self.result(content=b""))

    def test_summary_only_payload_is_failure(self):
        self.assert_rejected(lambda: self.result(filename="result_summary.txt"))

    def test_empty_compressed_log_is_failure(self):
        self.assert_rejected(lambda: self.result(content=gzip.compress(b""), filename="observer.log.20261021150100000.gz"))

    def test_missing_node_is_failure(self):
        with open(os.path.join(self.work, "nodes.json"), "w") as stream:
            json.dump([{"namespace": "test", "pod_name": "ob-pod"}, {"namespace": "test", "pod_name": "ob-other"}], stream)
        self.assert_rejected(self.result)

    def test_duplicate_node_is_failure(self):
        def result():
            payload = self.result()
            row = [self.node, "Success", "12.0 KB", "file save in " + payload]
            self.result([row, row])
        self.assert_rejected(result)

    def test_corrupt_archive_is_failure(self):
        self.assert_rejected(lambda: self.result(corrupt=True), tarfile.ReadError)

    def test_atomic_publish_failure_removes_temporary_and_credentials(self):
        with self.assertRaises(OSError):
            self.execute(self.result, OSError("simulated disk failure"))
        self.assertEqual(os.listdir(self.temp.name), [])

    def test_atomic_publish_does_not_expose_partial_attachment(self):
        original = os.replace
        def publish(source, destination):
            self.assertFalse(os.path.exists(destination))
            with tarfile.open(source, "r:gz") as archive:
                self.assertEqual(len(archive.getnames()), 2)
            original(source, destination)
        self.execute(self.result, publish)

    def test_cleanup_failure_does_not_publish(self):
        original = collector.shutil.rmtree
        def cleanup(path, **kwargs):
            if not kwargs.get("ignore_errors"):
                raise OSError("simulated cleanup failure")
            return original(path, **kwargs)
        with mock.patch.object(collector.shutil, "rmtree", side_effect=cleanup):
            with self.assertRaises(OSError):
                self.execute(self.result)
        self.assertFalse(os.path.exists(self.attachment))
        self.assertEqual(os.listdir(self.temp.name), [])

    def test_nonzero_command_cleans_credentials(self):
        with mock.patch.object(collector.subprocess, "call", return_value=2):
            with self.assertRaises(collector.CollectionError):
                collector.collect("start", "end", self.work, self.attachment)
        self.assertFalse(os.path.exists(self.work))
        self.assertFalse(os.path.exists(self.attachment))


if __name__ == "__main__":
    unittest.main()

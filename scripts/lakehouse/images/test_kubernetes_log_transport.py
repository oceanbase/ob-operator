import base64
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import sys
import tarfile
import tempfile
import types
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location("transport", Path(__file__).with_name("kubernetes_log_transport.py"))
transport = importlib.util.module_from_spec(spec)
spec.loader.exec_module(transport)


def wire(data):
    return (json.dumps({"size": len(data)}) + "\n" + base64.b64encode(data).decode() + "\n" +
            json.dumps({"size": len(data), "sha256": hashlib.sha256(data).hexdigest()}) + "\n")


class WebSocket:
    def __init__(self, chunks, status=0, stuck=False):
        self.chunks = list(chunks)
        self.output = []
        self.returncode = status
        self.open = True
        self.stuck = stuck
        self.closed = False

    def is_open(self): return self.open
    def update(self, timeout):
        if self.chunks:
            self.output.append(self.chunks.pop(0))
        elif not self.stuck:
            self.open = False
    def peek_stdout(self): return bool(self.output)
    def read_stdout(self): return self.output.pop(0)
    def peek_stderr(self): return False
    def close(self): self.closed = True


class TransportTest(unittest.TestCase):
    def setUp(self):
        self.client = types.SimpleNamespace(client=types.SimpleNamespace(connect_get_namespaced_pod_exec=object()),
            pod_name="test-ob", namespace="test", container_name="observer")

    def test_python_probe_is_memory_only(self):
        with tempfile.TemporaryDirectory() as directory:
            subprocess.run([sys.executable, "-c", transport.PROBE_CODE], cwd=directory, check=True)
            self.assertEqual(os.listdir(directory), [])

    def test_remote_python_pack_and_download_roundtrip(self):
        with tempfile.TemporaryDirectory(prefix="log space ' ") as directory:
            source = os.path.join(directory, "logs")
            os.mkdir(source)
            data = b"diagnostic data\x00\xff" * 9000
            with open(os.path.join(source, "observer.log"), "wb") as output:
                output.write(data)
            destination = source + ".tar.gz"
            subprocess.run([sys.executable, "-c", transport.PACK_CODE, source, destination], check=True)
            raw = subprocess.check_output([sys.executable, "-c", transport.DOWNLOAD_CODE, destination])
            received = io.BytesIO()
            decoder = transport.Decoder(received)
            for offset in range(0, len(raw), 7):
                decoder.feed(raw[offset:offset+7])
            decoder.finish()
            with tarfile.open(fileobj=io.BytesIO(received.getvalue()), mode="r:gz") as archive:
                self.assertEqual(archive.extractfile("logs/observer.log").read(), data)

    def test_pack_rejects_symlink_without_archive(self):
        with tempfile.TemporaryDirectory() as directory:
            source = os.path.join(directory, "logs")
            os.mkdir(source)
            os.symlink("/etc/passwd", os.path.join(source, "observer.log"))
            result = subprocess.run([sys.executable, "-c", transport.PACK_CODE, source, source+".tar.gz"],
                stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            self.assertNotEqual(result.returncode, 0)
            self.assertFalse(os.path.exists(source+".tar.gz"))

    def test_download_rejects_remote_symlink(self):
        with tempfile.TemporaryDirectory() as directory:
            path = os.path.join(directory, "archive")
            os.symlink("/etc/passwd", path)
            result = subprocess.run([sys.executable, "-c", transport.DOWNLOAD_CODE, path],
                stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(result.stdout, b"")

    def test_decoder_rejects_truncation_checksum_size_and_trailing_data(self):
        good = wire(b"test archive")
        for bad in (good[:-2], good.replace(hashlib.sha256(b"test archive").hexdigest(), "0"*64),
                    good.replace('"size": 12', '"size": 13', 1), good+"unexpected\n"):
            with self.subTest(value=bad):
                decoder = transport.Decoder(io.BytesIO())
                with self.assertRaises((transport.TransportError, ValueError)):
                    decoder.feed(bad)
                    decoder.finish()

    def test_decoder_rejects_invalid_frames_and_size_limits(self):
        for bad in ('{"size":0}\n', '{"size":2147483649}\n', '{"size":true}\n',
                    '{"size":1}\n!!!!\n', '{"size":1}\nYWI=\n', 'x'*65537):
            with self.subTest(value=bad[:40]):
                with self.assertRaises((transport.TransportError, ValueError)):
                    transport.Decoder(io.BytesIO()).feed(bad)

    def test_exec_requires_exit_status_and_closes_stream(self):
        for status in (1, None):
            websocket = WebSocket(["output"], status=status)
            with self.assertRaises(transport.TransportError):
                transport.execute(self.client, ["test"], stream_factory=lambda *args, **kwargs: websocket)
            self.assertTrue(websocket.closed)
        websocket = WebSocket(["out", "put"])
        self.assertEqual(transport.execute(self.client, ["test"],
            stream_factory=lambda *args, **kwargs: websocket), "output")
        self.assertTrue(websocket.closed)

    def test_exec_timeout_and_output_limit(self):
        websocket = WebSocket([], stuck=True)
        with mock.patch.object(transport.time, "monotonic", side_effect=[0, 181]):
            with self.assertRaisesRegex(transport.TransportError, "timed out"):
                transport.execute(self.client, ["test"], stream_factory=lambda *args, **kwargs: websocket)
        self.assertTrue(websocket.closed)
        with self.assertRaises(transport.TransportError):
            transport.execute(self.client, ["test"], stream_factory=lambda *args, **kwargs: WebSocket(["x"*16385]))

    def test_closed_stream_buffer_drain_and_final_sink_obey_deadline(self):
        for count in (1, 10):
            clock = [0]
            websocket = WebSocket([])
            websocket.open = False
            websocket.output = ["data"] * count
            def sink(block):
                clock[0] += 200 if count == 1 else 20
            with mock.patch.object(transport.time, "monotonic", side_effect=lambda: clock[0]):
                with self.assertRaisesRegex(transport.TransportError, "timed out"):
                    transport.execute(self.client, ["test"], sink=sink,
                        stream_factory=lambda *args, **kwargs: websocket)
            self.assertTrue(websocket.closed)

    def test_pack_prefers_tar_and_python_fallback_is_explicit(self):
        self.client._lakehouse_python = "python3"
        for capability in ("yes", ""):
            with mock.patch.object(transport, "execute", side_effect=[capability, ""]) as execute:
                transport.archive_logs(self.client, "/tmp/logs", "/tmp/logs.tar.gz")
                args = execute.call_args_list[-1][0][1]
                self.assertEqual(args[0], "tar" if capability else "python3")
        with mock.patch.object(transport, "execute", side_effect=["yes", transport.TransportError("tar failed")]):
            with self.assertRaises(transport.TransportError):
                transport.archive_logs(self.client, "/tmp/logs", "/tmp/logs.tar.gz")

    def test_download_publishes_atomically_and_cleans_failure(self):
        self.client._lakehouse_python = "python3"
        with tempfile.TemporaryDirectory() as directory:
            destination = os.path.join(directory, "archive.tar.gz")
            def deliver(client, command, sink, **kwargs):
                self.assertFalse(os.path.exists(destination))
                self.assertEqual(command[-1], "/tmp/logs.tar.gz")
                sink(wire(b"complete archive"))
            with mock.patch.object(transport, "execute", side_effect=deliver):
                self.assertTrue(transport.download(self.client, "/tmp/logs.tar.gz", destination))
            self.assertEqual(Path(destination).read_bytes(), b"complete archive")
            self.assertEqual(os.stat(destination).st_mode & 0o777, 0o600)
            def fail(client, command, sink, **kwargs):
                sink(wire(b"replacement"))
                raise transport.TransportError("missing exit status")
            with mock.patch.object(transport, "execute", side_effect=fail):
                with self.assertRaises(transport.TransportError):
                    transport.download(self.client, "/tmp/logs.tar.gz", destination)
            self.assertEqual(Path(destination).read_bytes(), b"complete archive")
            self.assertEqual(os.listdir(directory), ["archive.tar.gz"])

    def test_paths_cannot_escape_or_publish_through_symlink(self):
        self.client._lakehouse_python = "python3"
        for path in ("relative", "/", "/tmp/../private", "/tmp/a\x00b"):
            with self.assertRaises(transport.TransportError):
                transport.archive_logs(self.client, path, path+".tar.gz")
        with tempfile.TemporaryDirectory() as directory:
            destination = os.path.join(directory, "archive")
            os.symlink("/etc/passwd", destination)
            with self.assertRaises(transport.TransportError):
                transport.download(self.client, "/tmp/archive", destination)


if __name__ == "__main__":
    unittest.main()

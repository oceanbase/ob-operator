"""Exercise the packaged shell gate without invoking a real installer/cluster."""
import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


ENTRY = Path(__file__).resolve().parents[2] / "install.sh"


@unittest.skipUnless(shutil.which("bash") and shutil.which("sha256sum"),
                     "Bash and Linux-compatible sha256sum are required")
class InstallEntryTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="lakehouse entry with spaces ")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.bundle = self.root / "extracted bundle"
        self.bundle.mkdir()
        self.other = self.root / "unrelated working directory"
        self.other.mkdir()
        self.trace = self.root / "installer arguments"
        shutil.copyfile(ENTRY, self.bundle / "install.sh")
        (self.bundle / "bin").mkdir()
        self.installer = self.bundle / "bin/lakehouse-installer"
        self.installer.write_text('#!/bin/sh\nprintf "%s\\0" "$@" > "$TEST_INSTALLER_TRACE"\n'
                                  'exit "${TEST_INSTALLER_EXIT:-0}"\n')
        self.installer.chmod(0o755)
        for name in ("images/minio-images.tar.gz", "charts/example.yaml", "image-lock.json"):
            path = self.bundle / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(b"fixture package data\n")
        self.seal()

    def seal(self):
        records = []
        for path in sorted(self.bundle.rglob("*")):
            if path.is_file() and path.name != "SHA256SUMS":
                records.append(hashlib.sha256(path.read_bytes()).hexdigest() + "  " +
                               path.relative_to(self.bundle).as_posix())
        (self.bundle / "SHA256SUMS").write_text("\n".join(records) + "\n")

    def run_entry(self, args=(), env=None):
        environment = dict(os.environ, TEST_INSTALLER_TRACE=str(self.trace))
        if env:
            environment.update(env)
        return subprocess.run([shutil.which("bash"), str(self.bundle / "install.sh")] + list(args),
                              cwd=str(self.other), env=environment, stdout=subprocess.PIPE,
                              stderr=subprocess.PIPE, universal_newlines=True, timeout=10)

    def assert_blocked(self, result, message):
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        self.assertIn(message, result.stderr)
        self.assertFalse(self.trace.exists(), "unverified installer executed")

    def test_success_verifies_bundle_and_forwards_exact_arguments(self):
        args = ["--release", "demo", "-f", "values path ' quoted.yaml", "--dashboard-host", "127.0.0.1"]
        result = self.run_entry(args)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn("Package integrity verified", result.stdout)
        actual = self.trace.read_bytes().decode().split("\0")[:-1]
        self.assertEqual(actual, ["--bundle", str(self.bundle)] + args)

    def test_installer_exit_status_is_preserved(self):
        result = self.run_entry(["--check"], {"TEST_INSTALLER_EXIT": "17"})
        self.assertEqual(result.returncode, 17)
        self.assertTrue(self.trace.exists())

    def test_corrupt_or_missing_packaged_files_prevent_execution(self):
        for name in ("images/minio-images.tar.gz", "charts/example.yaml", "image-lock.json", "bin/lakehouse-installer"):
            path = self.bundle / name
            original = path.read_bytes()
            for action in ("corrupt", "remove"):
                with self.subTest(name=name, action=action):
                    if action == "corrupt":
                        path.write_bytes(original + b"modified")
                    else:
                        path.unlink()
                    self.assert_blocked(self.run_entry(), "Package verification failed")
                    path.write_bytes(original)
                    if name == "bin/lakehouse-installer":
                        path.chmod(0o755)

    def test_missing_manifest_prevents_execution(self):
        (self.bundle / "SHA256SUMS").unlink()
        self.assert_blocked(self.run_entry(), "Package checksum manifest is missing")

    def test_missing_checksum_tool_prevents_execution(self):
        commands = self.root / "only dirname available"
        commands.mkdir()
        (commands / "dirname").symlink_to(shutil.which("dirname"))
        self.assert_blocked(self.run_entry(env={"PATH": str(commands)}),
                            "Package verification requires sha256sum")

    def test_non_executable_installer_prevents_execution(self):
        self.installer.chmod(0o644)
        self.assert_blocked(self.run_entry(), "Package installer is missing or not executable")


if __name__ == "__main__":
    unittest.main()

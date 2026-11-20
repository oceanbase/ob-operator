import json
from pathlib import Path
import stat
import subprocess
import sys
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[1] / "configure-docker.py"
MIRRORS = [
    "https://mirror.ccs.tencentyun.com",
    "https://registry.docker-cn.com",
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://docker.1panel.live",
    "https://docker.1ms.run",
    "https://dytt.online",
    "https://lispy.org",
    "https://docker.xiaogenban1993.com",
    "https://docker.yomansunter.com",
    "https://666860.xyz",
    "https://a.ussh.net",
    "https://hub.rat.dev",
    "https://docker.m.daocloud.io",
]


class ConfigureDockerTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.path = Path(self.directory.name) / "daemon.json"

    def run_helper(self, *arguments):
        return subprocess.run(
            [sys.executable, str(SCRIPT), "--config", str(self.path), *arguments],
            text=True, capture_output=True, check=False,
        )

    def test_preview_does_not_write_existing_or_missing_config(self):
        result = self.run_helper()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads(result.stdout)["registry-mirrors"], MIRRORS)
        self.assertEqual(list(self.path.parent.iterdir()), [])
        original = '{ "data-root": "/custom/docker" }\n'
        self.path.write_text(original)
        result = self.run_helper()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.path.read_text(), original)
        self.assertEqual(list(self.path.parent.iterdir()), [self.path])

    def test_apply_preserves_settings_and_exact_backup(self):
        original = json.dumps({
            "data-root": "/custom/docker",
            "log-opts": {"max-size": "10m"},
            "registry-mirrors": ["https://company.example"],
            "max-concurrent-downloads": 2,
            "max-concurrent-uploads": 1,
        }) + "\n"
        self.path.write_text(original)
        self.path.chmod(0o640)
        result = self.run_helper("--apply")
        self.assertEqual(result.returncode, 0, result.stderr)
        config = json.loads(self.path.read_text())
        expected = json.loads(original)
        expected["registry-mirrors"].extend(MIRRORS)
        self.assertEqual(config, expected)
        self.assertEqual(stat.S_IMODE(self.path.stat().st_mode), 0o640)
        backups = list(self.path.parent.glob("daemon.json.backup-*"))
        self.assertEqual(len(backups), 1)
        self.assertEqual(backups[0].read_text(), original)
        self.assertEqual(self.run_helper("--apply").returncode, 0)
        self.assertEqual(len(list(self.path.parent.glob("daemon.json.backup-*"))), 1)

    def test_custom_mirrors_replace_sample_and_deduplicate(self):
        self.path.write_text('{"registry-mirrors":["https://company.example/"]}')
        result = self.run_helper("--apply", "--mirror", "https://company.example",
                                 "--mirror", "https://second.example/")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads(self.path.read_text())["registry-mirrors"],
                         ["https://company.example/", "https://second.example"])

    def test_new_config_has_defaults_and_no_backup(self):
        result = self.run_helper("--apply")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads(self.path.read_text()), {
            "registry-mirrors": MIRRORS,
            "max-concurrent-downloads": 10,
            "max-concurrent-uploads": 5,
        })
        self.assertEqual(json.loads(self.path.read_text()), json.loads(
            SCRIPT.with_name("docker-daemon.json").read_text()))
        self.assertEqual(list(self.path.parent.iterdir()), [self.path])

    def test_invalid_config_leaves_original_untouched(self):
        for content in ("{", "[]", '{"registry-mirrors":null}',
                        '{"registry-mirrors":[1]}', '{"max-concurrent-downloads":true}',
                        '{"max-concurrent-uploads":0}', '{"debug":true,"debug":false}'):
            with self.subTest(content=content):
                self.path.write_text(content)
                result = self.run_helper("--apply")
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(self.path.read_text(), content)
                self.assertEqual(list(self.path.parent.iterdir()), [self.path])

    def test_symlink_is_rejected(self):
        target = self.path.parent / "original.json"
        target.write_text("{}")
        self.path.symlink_to(target)
        result = self.run_helper("--apply")
        self.assertNotEqual(result.returncode, 0)
        self.assertTrue(self.path.is_symlink())
        self.assertEqual(target.read_text(), "{}")
        self.assertEqual(len(list(self.path.parent.iterdir())), 2)

    def test_invalid_new_mirror_never_creates_config(self):
        for mirror in ("http://example.com", "https://user:secret@example.com",
                       "https://example.com/?token=value", "https://example.com/#fragment"):
            with self.subTest(mirror=mirror):
                result = self.run_helper("--apply", "--mirror", mirror)
                self.assertNotEqual(result.returncode, 0)
                self.assertFalse(self.path.exists())


if __name__ == "__main__":
    unittest.main()

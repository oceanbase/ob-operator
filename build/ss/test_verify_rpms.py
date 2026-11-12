import hashlib
import importlib.util
import tempfile
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location('verify_rpms', Path(__file__).with_name('verify-rpms.py'))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class VerifyRPMs(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / 'main.rpm').write_bytes(b'original')
        self.entry = hashlib.sha256(b'original').hexdigest() + '  main.rpm\n'
        (self.root / 'SHA256SUMS').write_text(self.entry)

    def reject(self):
        with self.assertRaises(ValueError):
            module.verify(self.root)

    def test_valid(self):
        module.verify(self.root)

    def test_unlisted_rpm(self):
        (self.root / 'extra.rpm').write_bytes(b'extra')
        self.reject()

    def test_missing_rpm(self):
        (self.root / 'main.rpm').unlink()
        self.reject()

    def test_corrupted_rpm(self):
        (self.root / 'main.rpm').write_bytes(b'changed')
        self.reject()

    def test_duplicate_entry(self):
        (self.root / 'SHA256SUMS').write_text(self.entry * 2)
        self.reject()

    def test_empty_manifest(self):
        (self.root / 'SHA256SUMS').write_text('')
        self.reject()

    def test_path_escape(self):
        (self.root / 'SHA256SUMS').write_text(self.entry.replace('main.rpm', '../main.rpm'))
        self.reject()

    def test_symlink(self):
        (self.root / 'main.rpm').rename(self.root / 'original')
        (self.root / 'main.rpm').symlink_to('original')
        self.reject()


if __name__ == '__main__':
    unittest.main()

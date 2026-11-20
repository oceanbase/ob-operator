"""Verify the actual Compose entrypoint normalizes inherited file limits."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


class ObserverFileLimits(unittest.TestCase):
    def test_children_inherit_hard_file_limit(self):
        entrypoint = Path(__file__).resolve().parent / 'scripts/observer-entrypoint.sh'
        with tempfile.TemporaryDirectory() as directory:
            # Stop at the first filesystem operation, before starting services.
            mkdir = Path(directory) / 'mkdir'
            mkdir.write_text('#!/bin/bash\nulimit -Sn\nexit 125\n')
            mkdir.chmod(0o755)
            env = dict(os.environ, PATH=directory + ':' + os.environ['PATH'])
            for soft in (1024, 65536):
                with self.subTest(soft=soft):
                    result = subprocess.run(
                        ['bash', '-c', 'ulimit -Sn "$1"; ulimit -Hn 65536; exec bash "$2"',
                         'test', str(soft), str(entrypoint)],
                        env=env, text=True, capture_output=True)
                    self.assertEqual(result.returncode, 125, result.stderr)
                    self.assertEqual(result.stdout.strip(), '65536')


if __name__ == '__main__':
    unittest.main()

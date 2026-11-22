"""Exercise route discovery without changing the host network or secret files."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPT = Path(os.environ.get('HOST_INFO_SCRIPT', Path(__file__).parent / 'scripts/host-info-init.sh'))


class HostInfoTests(unittest.TestCase):
    def discover(self, route='', override='', status=0, probe='1.1.1.1'):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        root = Path(temporary.name)
        target = root / 'docker_host_ip'
        script = root / 'host-info-init.sh'
        script.write_text(SCRIPT.read_text().replace('/run/secrets/docker_host_ip', str(target)))
        ip = root / 'ip'
        ip.write_text('#!/bin/sh\nprintf "%s\\n" "$@" > "$IP_ARGS"\nprintf "%s\\n" "$ROUTE"\nexit "$IP_STATUS"\n')
        ip.chmod(0o755)
        env = dict(os.environ, PATH=str(root) + ':' + os.environ['PATH'],
                   HOST_ACCESS_IP=override, HOST_IP_PROBE=probe, ROUTE=route,
                   IP_STATUS=str(status), IP_ARGS=str(root / 'args'))
        result = subprocess.run(['sh', str(script)], env=env, text=True, capture_output=True)
        return root, target, result

    def test_route_sources(self):
        for route, expected in (
            ('1.1.1.1 via 192.0.2.1 dev eth0 src 192.0.2.4 uid 0\n    cache', '192.0.2.4'),
            ('2001:db8::1 dev eth0 src 2001:db8::4 metric 1', '2001:db8::4'),
            ('1.1.1.1 dev eth0\n    src 192.0.2.5 uid 0', '192.0.2.5'),
        ):
            with self.subTest(route=route):
                root, target, result = self.discover(route=route, probe='192.0.2.1')
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(target.read_text(), expected + '\n')
                self.assertEqual(target.stat().st_mode & 0o777, 0o600)
                self.assertEqual((root / 'args').read_text().splitlines(), ['route', 'get', '192.0.2.1'])
                self.assertEqual(list(root.glob('docker_host_ip.tmp.*')), [])

    def test_explicit_ip_skips_route_probe(self):
        root, target, result = self.discover(override='2001:db8::4', status=99)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(target.read_text(), '2001:db8::4\n')
        self.assertFalse((root / 'args').exists())

    def test_invalid_or_missing_source_fails_without_writing(self):
        for route, override, status in (
            ('1.1.1.1 dev eth0', '', 0), ('', '', 2),
            ('1.1.1.1 dev eth0 src', '', 0),
            ('1.1.1.1 dev eth0 src not-an-ip', '', 0),
            ('', 'host;touch SHOULD_NOT_EXIST', 0),
        ):
            with self.subTest(route=route, override=override, status=status):
                _, target, result = self.discover(route, override, status)
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn('Unable to determine Docker host IP', result.stderr)
                self.assertFalse(target.exists())


if __name__ == '__main__':
    unittest.main()

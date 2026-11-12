"""Execute Compose access notes and check the argv of their printed commands."""
import os
from pathlib import Path
import shlex
import subprocess
import tempfile
import unittest

SCRIPT = Path(os.environ.get('ACCESS_NOTES_SCRIPT', Path(__file__).parent / 'scripts/observer-bootstrap.sh'))


class AccessNotesTests(unittest.TestCase):
    def render(self, host='192.0.2.4', project=None, args=(), failure=False):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        secrets = root / 'secrets'
        secrets.mkdir()
        (secrets / 'docker_host_ip').write_text(host)
        bootstrap = root / 'bootstrap'
        bootstrap.write_text('#!/bin/sh\nexit ' + ('23' if failure else '0') + '\n')
        bootstrap.chmod(0o755)
        script = root / 'wrapper.sh'
        script.write_text(SCRIPT.read_text().replace('/usr/local/bin/lakehouse-bootstrap', str(bootstrap)).replace('/run/secrets', str(secrets)))
        env = dict(os.environ)
        env.pop('COMPOSE_PROJECT_NAME', None)
        if project is not None:
            env['COMPOSE_PROJECT_NAME'] = project
        env.update(OBSERVER_SQL_HOST_PORT='33881', OBSHELL_HOST_PORT='33886')
        result = subprocess.run(['sh', str(script), *args], env=env, text=True, capture_output=True)
        return root, result

    def test_default_commands_are_complete(self):
        root, result = self.render()
        self.assertEqual(result.returncode, 0, result.stderr)
        sql = next(line.strip() for line in result.stdout.splitlines() if line.strip().startswith('mysql '))
        credentials = next(line.strip() for line in result.stdout.splitlines() if line.strip().startswith('sudo docker compose '))
        self.assertEqual(shlex.split(sql), ['mysql', '-h', '192.0.2.4', '-P', '33881', '-u', 'root@sys', '-p'])
        self.assertEqual(shlex.split(credentials), ['sudo', 'docker', 'compose', '-p', 'ob-lakehouse', 'run', '--rm', '--no-deps', 'observer-bootstrap', '--mode', 'show-password'])
        self.assertIn('http://192.0.2.4:33886/', result.stdout)
        self.assertEqual((root / 'secrets/observer_bootstrap_complete').stat().st_mode & 0o777, 0o600)

    def test_custom_project_is_preserved(self):
        _, result = self.render(project='notes-second-project')
        command = next(line.strip() for line in result.stdout.splitlines() if line.strip().startswith('sudo docker compose '))
        self.assertEqual(shlex.split(command)[4], 'notes-second-project')
        self.assertIn('same options before run', result.stdout)

    def test_ipv6_and_shell_quoting(self):
        _, result = self.render(host='2001:db8::12')
        self.assertIn('http://[2001:db8::12]:33886/', result.stdout)
        host = "host'$(touch SHOULD_NOT_EXIST)"
        root, result = self.render(host=host)
        command = next(line.strip() for line in result.stdout.splitlines() if line.strip().startswith('mysql '))
        stub = root / 'mysql'
        stub.write_text('#!/bin/sh\nprintf "%s\\n" "$@" > "$ARGS_FILE"\n')
        stub.chmod(0o755)
        argv_file = root / 'args'
        subprocess.run(['sh', '-c', command], cwd=root, env=dict(os.environ, PATH=str(root)+':'+os.environ['PATH'], ARGS_FILE=str(argv_file)), check=True)
        self.assertEqual(argv_file.read_text().splitlines()[1], host)
        self.assertFalse((root / 'SHOULD_NOT_EXIST').exists())

    def test_password_mode_does_not_print_success_notes(self):
        for args in [('--mode', 'show-password'), ('--mode=show-password',)]:
            with self.subTest(args=args):
                root, result = self.render(args=args)
                self.assertEqual(result.returncode, 0)
                self.assertNotIn('deployment completed', result.stdout)
                self.assertFalse((root / 'secrets/observer_bootstrap_complete').exists())

    def test_failed_bootstrap_has_no_success_notes(self):
        root, result = self.render(failure=True)
        self.assertEqual(result.returncode, 23)
        self.assertNotIn('deployment completed', result.stdout)
        self.assertFalse((root / 'secrets/observer_bootstrap_complete').exists())


if __name__ == '__main__':
    unittest.main()

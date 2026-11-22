"""Run the real supervisor against an isolated, signal-aware daemon fixture."""
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import time
import unittest


SCRIPT = Path(os.environ.get('LOGSERVICE_SCRIPT', Path(__file__).parent / 'scripts/logservice-entrypoint.sh'))


class LogServiceEntrypointTests(unittest.TestCase):
    def run_supervisor(self, stop_signal=None):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            home = root / 'home'
            (home / 'bin').mkdir(parents=True)
            (home / 'etc').mkdir()
            (home / 'store/etc').mkdir(parents=True)
            (home / 'store/etc/config.toml').write_text('retained fixture configuration')
            proc = root / 'proc'
            proc.mkdir()
            (proc / 'self').symlink_to('/proc/self')
            daemon = root / 'daemon.py'
            daemon.write_text('''import ctypes, os, signal, time
from pathlib import Path
root = Path(os.environ['FIXTURE_ROOT'])
ctypes.CDLL(None).prctl(15, b'oblogservice', 0, 0, 0)
def stop(sig, frame):
    (root / 'daemon-signal').write_text(str(sig))
    raise SystemExit(0)
signal.signal(signal.SIGTERM, stop)
(root / 'proc' / str(os.getpid())).symlink_to('/proc/' + str(os.getpid()))
(root / 'daemon-pid').write_text(str(os.getpid()))
while True:
    time.sleep(0.05)
''')
            launcher = home / 'bin/oblogservice'
            launcher.write_text('#!/bin/sh\npython3 "$FIXTURE_ROOT/daemon.py" </dev/null >/dev/null 2>&1 &\n')
            launcher.chmod(0o755)
            sleeper = root / 'sleep'
            sleeper.write_text('#!/bin/sh\nif [ "$1" = 5 ]; then touch "$FIXTURE_ROOT/supervising"; fi\nexec /bin/sleep "$@"\n')
            sleeper.chmod(0o755)
            script = root / 'entrypoint.sh'
            script.write_text(SCRIPT.read_text().replace('/home/admin/oblogservice', str(home)).replace('/proc/', str(proc) + '/'))
            env = dict(os.environ, FIXTURE_ROOT=str(root), PATH=str(root) + ':' + os.environ['PATH'])
            daemon_pid = None
            with (root / 'stdout').open('w') as stdout, (root / 'stderr').open('w') as stderr:
                supervisor = subprocess.Popen(['bash', str(script)], env=env, stdout=stdout, stderr=stderr, start_new_session=True)
                try:
                    deadline = time.monotonic() + 10
                    while not (root / 'supervising').exists() and time.monotonic() < deadline:
                        if supervisor.poll() is not None:
                            self.fail('supervisor exited before startup: ' + (root / 'stderr').read_text())
                        time.sleep(0.05)
                    self.assertTrue((root / 'supervising').exists(), 'supervisor never became ready')
                    daemon_pid = int((root / 'daemon-pid').read_text())
                    if stop_signal is None:
                        os.kill(daemon_pid, signal.SIGTERM)
                    else:
                        supervisor.send_signal(stop_signal)
                    status = supervisor.wait(timeout=10)
                    self.assertEqual(status, 1 if stop_signal is None else 0, (root / 'stderr').read_text())
                    self.assertEqual((root / 'daemon-signal').read_text(), str(int(signal.SIGTERM)))
                    expected = 'daemon exited' if stop_signal is None else 'Stopping LogService pid'
                    self.assertIn(expected, (root / 'stderr').read_text())
                    self.assertNotIn('SIGKILL', (root / 'stderr').read_text())
                finally:
                    try:
                        os.killpg(supervisor.pid, signal.SIGKILL)
                    except ProcessLookupError:
                        pass
                    supervisor.wait(timeout=5)

    def test_sigterm_stops_daemon(self):
        self.run_supervisor(signal.SIGTERM)

    def test_sigint_stops_daemon(self):
        self.run_supervisor(signal.SIGINT)

    def test_unexpected_daemon_exit_fails_supervisor(self):
        self.run_supervisor()


if __name__ == '__main__':
    unittest.main()

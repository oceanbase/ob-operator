"""Exercise the current Python prerequisite runner's shared deadline contract."""
import importlib.util
from pathlib import Path
import unittest
from unittest import mock

SPEC = importlib.util.spec_from_file_location('bootstrap', Path(__file__).parent / 'prerequisites/bootstrap.py')
BOOTSTRAP = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(BOOTSTRAP)

class TimeoutTests(unittest.TestCase):
    def test_operator_wait_uses_remaining_shared_budget(self):
        for seconds in (1800, 1200):
            with self.subTest(seconds=seconds), mock.patch.object(BOOTSTRAP.time, 'monotonic', return_value=100):
                kube = mock.Mock()
                BOOTSTRAP.wait(kube, 100 + seconds, 'rollout', 'status', 'deployment/operator')
                kube.run.assert_called_once_with('rollout', 'status', 'deployment/operator', '--timeout=%ds' % seconds, timeout=seconds)

    def test_expired_budget_never_starts_another_wait(self):
        with mock.patch.object(BOOTSTRAP.time, 'monotonic', return_value=101):
            kube = mock.Mock()
            with self.assertRaises(BOOTSTRAP.BootstrapError):
                BOOTSTRAP.wait(kube, 100, 'rollout', 'status', 'deployment/operator')
            kube.run.assert_not_called()

if __name__ == '__main__':
    unittest.main()

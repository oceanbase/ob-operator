import argparse
import copy
import importlib.util
from pathlib import Path
import unittest
import os
import shutil
import subprocess
import tempfile
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('progress', Path(__file__).with_name('progress.py'))
p = importlib.util.module_from_spec(spec)
spec.loader.exec_module(p)

def obj(kind, name, status=None, owner=None, labels=None):
    return {'kind': kind, 'metadata': {'name': name, 'uid': name + '-uid', 'labels': labels or {},
            'ownerReferences': [{'uid': owner + '-uid', 'kind': 'Job' if owner.startswith('image') else 'OBCluster'}] if owner else []},
            'status': status or {}}

class ProgressTests(unittest.TestCase):
    def setUp(self):
        self.helm = {'version': 1, 'info': {'status': 'pending-install', 'last_deployed': '2026-11-04T09:37:14Z'}}
        self.ls = obj('OBLogServiceCluster', 'ob-demo-ls', {'status': 'running'})
        self.ob = obj('OBCluster', 'ob-demo-ob', {'status': 'new'})

    def launcher(self, fallback=False):
        with tempfile.TemporaryDirectory() as tmp:
            bindir = Path(tmp)
            old = bindir / 'python3'
            old.write_text('#!/bin/sh\nif [ "$1" = "-c" ]; then exit 1; fi\necho OLD_RUNTIME_SELECTED\nexit 97\n')
            old.chmod(0o755)
            (bindir / 'dirname').symlink_to(shutil.which('dirname'))
            if fallback:
                newer = bindir / 'python3.9'
                newer.write_text('#!/bin/sh\nif [ "$1" = "-c" ]; then exit 0; fi\nprintf "[%s]\\n" "$@"\n')
                newer.chmod(0o755)
            env = dict(os.environ, PATH=tmp)
            return subprocess.run(['/bin/sh', str(Path(__file__).with_name('progress.sh')),
                                   '--namespace', 'namespace with spaces', '--once'],
                                  env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                  universal_newlines=True)

    def test_launcher_rejects_unsupported_default_python_without_traceback(self):
        result = self.launcher()
        self.assertEqual(result.returncode, 2)
        self.assertIn('Python 3.8+', result.stderr)
        self.assertNotIn('OLD_RUNTIME_SELECTED', result.stdout)

    def test_launcher_selects_supported_python_and_preserves_arguments(self):
        result = self.launcher(fallback=True)
        self.assertEqual(result.returncode, 0)
        self.assertIn('[namespace with spaces]', result.stdout)
        self.assertIn('[--once]', result.stdout)
        self.assertNotIn('OLD_RUNTIME_SELECTED', result.stdout)

    def test_kubeconfig_precedence_and_multi_file_merge(self):
        with tempfile.TemporaryDirectory() as tmp:
            home = Path(tmp)
            default = home / '.kube/config'
            default.parent.mkdir()
            default.write_text('default')
            explicit = home / 'explicit'
            explicit.write_text('explicit')
            merged = str(home / 'missing') + os.pathsep + str(explicit)
            with patch.object(p.Path, 'home', return_value=home), patch.dict(os.environ, {'KUBECONFIG': merged}):
                self.assertEqual(p.default_kubeconfig(), merged)
                explicit.unlink()
                self.assertEqual(p.default_kubeconfig(), str(default))

    def test_kubeconfig_readable_k3s_fallback_and_no_access(self):
        with patch.dict(os.environ, {'KUBECONFIG': ''}), patch.object(p.Path, 'is_file', return_value=True), patch.object(p.os, 'access', side_effect=lambda path, mode: str(path) == '/etc/rancher/k3s/k3s.yaml'):
            self.assertEqual(p.default_kubeconfig(), '/etc/rancher/k3s/k3s.yaml')
        with patch.object(p.os, 'access', return_value=False):
            with self.assertRaisesRegex(RuntimeError, 'private copy.*0600'):
                p.default_kubeconfig()

    def test_snapshot_preserves_environment_kubeconfig_merge(self):
        args = argparse.Namespace(context=None, kubeconfig=None, namespace='ob-demo', release='ob-demo')
        with tempfile.TemporaryDirectory() as tmp:
            first, second = Path(tmp) / 'cluster', Path(tmp) / 'credentials'
            first.touch(); second.touch()
            merged = str(first) + os.pathsep + str(second)
            with patch.dict(os.environ, {'KUBECONFIG': merged}), patch.object(p, 'execute', side_effect=[self.helm, {'items': []}, self.ls, self.ob]) as run:
                p.snapshot(args)
                for call in run.call_args_list:
                    self.assertNotIn('--kubeconfig', call.args[0])
                self.assertEqual(os.environ['KUBECONFIG'], merged)

    def test_helm_nanosecond_timestamp(self):
        self.assertIsNotNone(p.timestamp('2026-11-04T17:37:14.00117628+08:00'))
        self.assertIsNone(p.timestamp(None))

    def test_fresh_missing_crds_waits_not_failed(self):
        state = p.analyze('ob-demo', {}, [], [])
        self.assertEqual(state['stage'], 0)
        self.assertFalse(state['issues'])
        self.assertFalse(any(state['passed']))

    def test_random_pull_job_ownership_and_unrelated_namespace_noise(self):
        job = obj('Job', 'image-pull-ready-abc', owner='ob-demo-ob')
        pod = obj('Pod', 'image-pull-ready-abc-pod', {'phase': 'Pending'}, owner=job['metadata']['name'])
        unrelated = obj('Pod', 'other-crash', {'containerStatuses': [{'state': {'waiting': {'reason': 'CrashLoopBackOff'}}}]})
        event = obj('Event', 'pulling'); event.update({'involvedObject': {'uid': pod['metadata']['uid']}, 'reason': 'Pulling', 'message': 'pulling observer'})
        state = p.analyze('ob-demo', self.helm, [pod, job, unrelated, event], [self.ls, self.ob])
        self.assertEqual(state['stage'], 3)
        self.assertIn('Observer', state['detail'])
        self.assertEqual(state['latest']['reason'], 'Pulling')
        self.assertEqual(state['issues'], [])

    def test_pull_failure_and_failed_job_visible(self):
        pod = obj('Pod', 'ob-demo-ob-pod', {'containerStatuses': [{'state': {'waiting': {'reason': 'ImagePullBackOff', 'message': 'denied'}}}]}, labels={'ref-obcluster': 'ob-demo-ob'})
        job = obj('Job', 'ob-demo-readiness', {'conditions': [{'type': 'Failed', 'status': 'True'}]})
        state = p.analyze('ob-demo', self.helm, [pod, job], [self.ls, self.ob])
        self.assertTrue(any('ImagePullBackOff' in x for x in state['issues']))
        self.assertTrue(any('Job failed' in x for x in state['issues']))

    def test_deleted_hooks_and_deployed_is_success(self):
        self.helm['info']['status'] = 'deployed'
        state = p.analyze('ob-demo', self.helm, [], [self.ls, self.ob])
        self.assertTrue(state['passed'][-1])
        self.assertTrue(state['passed'][0])

    def test_upgrade_preflight_takes_priority_over_running_crs(self):
        self.ob['status']['status'] = 'running'
        self.helm['version'] = 2
        self.helm['info']['status'] = 'pending-upgrade'
        state = p.analyze('ob-demo', self.helm, [obj('Job', 'ob-demo-prerequisites')], [self.ls, self.ob])
        self.assertEqual(state['stage'], 0)
        self.assertFalse(state['passed'][0])

    def test_logservice_counts_use_operator_labels(self):
        pods = [obj('Pod', 'ob-demo-ls-zone' + str(i), {'conditions': [{'type': 'Ready', 'status': 'True'}]}, labels={'ref-oblogservicecluster': 'ob-demo-ls'}) for i in range(3)]
        state = p.analyze('ob-demo', self.helm, pods, [self.ls, self.ob])
        self.assertEqual(state['counts']['LogService'], '3/3 Pods ready')

    def test_forbidden_is_not_silently_treated_as_missing(self):
        args = argparse.Namespace(context=None, kubeconfig='/test/config', namespace='ob-demo', release='ob-demo')
        with patch.object(p, 'execute', side_effect=[self.helm, {'items': []}, RuntimeError('Forbidden')]):
            with self.assertRaisesRegex(RuntimeError, 'Forbidden'):
                p.snapshot(args)

    def test_snapshot_is_readonly_and_context_consistent(self):
        args = argparse.Namespace(context='test', kubeconfig='/tmp/config', namespace='ob-demo', release='ob-demo')
        with patch.object(p, 'execute', side_effect=[self.helm, {'items': []}, RuntimeError('(NotFound)'), RuntimeError("server doesn't have a resource type")]) as run:
            state = p.snapshot(args)
            self.assertEqual(state['stage'], 0)
            for call in run.call_args_list:
                cmd = call.args[0]
                self.assertIn('/tmp/config', cmd)
                self.assertIn('test', cmd)
                self.assertTrue('status' in cmd or 'get' in cmd)
                self.assertFalse(any(x in cmd for x in ['apply', 'delete', 'patch', 'upgrade', 'secrets']))

if __name__ == '__main__':
    unittest.main()

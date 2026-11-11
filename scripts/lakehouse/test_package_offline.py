"""Offline package regressions. Fixtures are not release image evidence."""
import argparse
import importlib.util
import json
import hashlib
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest
from unittest import mock

SPEC = importlib.util.spec_from_file_location('offline', Path(__file__).with_name('package-offline.py'))
OFFLINE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(OFFLINE)


class OfflineTests(unittest.TestCase):
    def test_compose_offline_config_keeps_project_resources_isolated(self):
        try:
            subprocess.run(['docker', 'compose', 'version'], check=True,
                           stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        except (OSError, subprocess.CalledProcessError):
            self.skipTest('Requires Compose CLI only; no daemon or image pull')
        with tempfile.TemporaryDirectory() as tmp:
            bundle = Path(tmp)
            (bundle / 'compose.yaml').write_text('''name: origin
services:
  smoke:
    image: busybox:1.36.1
    command: [echo, "$$HOME"]
    volumes: [data:/data, fixed:/fixed, shared:/shared]
    networks: [default, fixed, external]
volumes:
  data: {}
  fixed: {name: deliberately-shared-data}
  shared: {external: true, name: external-data}
networks:
  default: {}
  fixed: {name: deliberately-shared-network}
  external: {external: true}
''')
            (bundle / '.env').write_text('')
            (bundle / 'image-lock.json').write_text(json.dumps({
                'delivery_channel': 'compose', 'delivery_version': 'fixture',
                'platform': 'linux/amd64', 'images': [{'reference': 'busybox:1.36.1'}]}))
            args = argparse.Namespace(bundle=bundle, values=[], docker='docker')
            selected = OFFLINE.plan(args)
            frozen = bundle / 'offline.yaml'
            frozen.write_text(OFFLINE.yaml.safe_dump(selected['_compose_config']))
            for project in ('offline-a', 'offline-b'):
                config = json.loads(OFFLINE.run([
                    'docker', 'compose', '-p', project, '-f', frozen,
                    'config', '--format', 'json']))
                self.assertEqual(config['volumes']['data']['name'], project + '_data')
                self.assertEqual(config['networks']['default']['name'], project + '_default')
                self.assertEqual(config['volumes']['fixed']['name'], 'deliberately-shared-data')
                self.assertEqual(config['volumes']['shared']['name'], 'external-data')
                self.assertEqual(config['networks']['fixed']['name'], 'deliberately-shared-network')
                self.assertEqual(config['networks']['external']['name'], 'external')
                self.assertEqual(config['services']['smoke']['command'], ['echo', '$$HOME'])

    def test_real_cached_image_export_and_compose_round_trip(self):
        try:
            actual, = json.loads(subprocess.check_output(
                ['docker', 'image', 'inspect', 'busybox:1.36.1'], stderr=subprocess.DEVNULL))
        except (OSError, subprocess.CalledProcessError):
            self.skipTest('Requires cached busybox:1.36.1; test never pulls images')
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            bundle = root / 'bundle'
            bundle.mkdir()
            (bundle / 'compose.yaml').write_text('services:\n  smoke:\n    image: busybox:1.36.1\n'
                                                '    command: ["echo", "$$HOME"]\n')
            (bundle / '.env').write_text('')
            item = {'role': 'fixture', 'reference': 'busybox:1.36.1', 'image_id': actual['Id']}
            (bundle / 'image-lock.json').write_text(json.dumps({
                'delivery_channel': 'compose', 'delivery_version': 'fixture',
                'platform': actual['Os'] + '/' + actual['Architecture'], 'images': [item]}))
            args = argparse.Namespace(bundle=bundle, values=[], docker='docker',
                                      output=root / 'offline', pull='never')
            OFFLINE.export(args, OFFLINE.plan(args))
            with tarfile.open(args.output / 'images.tar.gz') as archive:
                manifest = json.load(archive.extractfile('manifest.json'))
                self.assertEqual(manifest[0]['RepoTags'], ['busybox:1.36.1'])
                raw = archive.extractfile(manifest[0]['Config']).read()
                self.assertEqual('sha256:' + hashlib.sha256(raw).hexdigest(), actual['Id'])
            for line in (args.output / 'SHA256SUMS').read_text().splitlines():
                digest, path = line.split('  ', 1)
                self.assertEqual(OFFLINE.sha256(args.output / path), digest)
            config = json.loads(OFFLINE.run(['docker', 'compose', '-f',
                                           args.output / 'deployment/compose.yaml',
                                           'config', '--format', 'json']))
            self.assertEqual(config['services']['smoke']['command'], ['echo', '$$HOME'])
            self.assertFalse((args.output / 'INCOMPLETE').exists())

    def test_bootstrap_multidocument_payload_images_are_included(self):
        payload = "apiVersion: apps/v1\nkind: Deployment\nspec:\n  template:\n    spec:\n      containers:\n        - image: cert-manager:1\n---\napiVersion: apps/v1\nkind: Deployment\nspec:\n  template:\n    spec:\n      containers:\n        - image: operator:2\n"
        documents = [{'kind': 'ConfigMap', 'data': {'manifests.yaml': payload,
                      'prepare.sh': '#!/bin/sh\necho ready'}}]
        self.assertEqual(OFFLINE.manifest_refs(documents, set()),
                         {'cert-manager:1', 'operator:2'})

    def test_dynamic_cr_and_diagnostic_images_are_included(self):
        docs = [
            {'kind': 'OBCluster', 'spec': {'observer': {'image': 'observer:1'},
                                         'topology': [{'observer': {'image': 'observer:2'}}]}},
            {'kind': 'OBLogServiceCluster', 'spec': {'logService': {'image': 'ls:1'}}},
            {'kind': 'ConfigMap', 'data': {'config.yaml':
                'inspection:\n  obdiag:\n    image: diag:1\nsqlAnalyzer:\n  image: analyzer:1\n'}},
            {'kind': 'Job', 'spec': {'template': {'spec': {
                'initContainers': [{'image': 'init:1'}], 'containers': [{'image': 'job:1'}]}}}},
        ]
        self.assertEqual(OFFLINE.manifest_refs(docs, {'analyzer:1'}),
                         {'observer:1', 'observer:2', 'ls:1', 'diag:1', 'init:1', 'job:1'})

    def test_inactive_exception_cannot_hide_real_pod(self):
        docs = [{'kind': 'Pod', 'spec': {'containers': [{'image': 'analyzer:1'}]}}]
        self.assertEqual(OFFLINE.manifest_refs(docs, {'analyzer:1'}), {'analyzer:1'})

    def test_config_identity_and_architecture_are_checked(self):
        item = {'reference': 'test:1', 'image_id': 'sha256:expected'}
        actual = {'Id': 'sha256:wrong', 'Os': 'linux', 'Architecture': 'amd64'}
        with mock.patch.object(OFFLINE, 'run', return_value=json.dumps([actual])):
            with self.assertRaisesRegex(ValueError, 'identity mismatch'):
                OFFLINE.inspect_image('docker', item, 'linux/amd64')
        actual['Id'] = item['image_id']
        actual['Architecture'] = 'arm64'
        with mock.patch.object(OFFLINE, 'run', return_value=json.dumps([actual])):
            with self.assertRaisesRegex(ValueError, 'platform mismatch'):
                OFFLINE.inspect_image('docker', item, 'linux/amd64')

    def test_unknown_compose_image_fails_closed(self):
        with tempfile.TemporaryDirectory() as tmp:
            bundle = Path(tmp)
            (bundle / 'image-lock.json').write_text(json.dumps({
                'delivery_channel': 'compose', 'images': []}))
            args = argparse.Namespace(bundle=bundle, values=[], docker='docker')
            with mock.patch.object(OFFLINE, 'run', return_value=json.dumps({
                    'services': {'unexpected': {'image': 'unexpected:1'}}})):
                with self.assertRaisesRegex(ValueError, 'outside this release lock'):
                    OFFLINE.plan(args)

    def test_failure_retains_incomplete_marker(self):
        with tempfile.TemporaryDirectory() as tmp:
            args = argparse.Namespace(bundle=Path(tmp) / 'bundle', output=Path(tmp) / 'output',
                                      docker='docker', pull='never')
            args.bundle.mkdir()
            selected = {'images': [{'reference': 'missing:1'}], 'platform': 'linux/amd64'}
            with mock.patch.object(OFFLINE, 'inspect_image', side_effect=ValueError('missing')):
                with self.assertRaises(ValueError):
                    OFFLINE.export(args, selected)
            self.assertTrue((args.output / 'INCOMPLETE').exists())
            self.assertFalse((args.output / 'INSTALL.txt').exists())

    def test_output_cannot_recursively_include_itself(self):
        with tempfile.TemporaryDirectory() as tmp:
            args = argparse.Namespace(bundle=Path(tmp), output=Path(tmp) / 'out')
            with self.assertRaisesRegex(ValueError, 'outside'):
                OFFLINE.export(args, {})


if __name__ == '__main__':
    unittest.main()

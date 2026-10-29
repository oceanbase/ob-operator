"""Native release identity and actual Helm/Compose packaging regressions.

Synthetic inspect records are explicitly metadata fixtures, never evidence that
release images were pulled, inspected on a live daemon, or installed.
"""
import argparse
import copy
import hashlib
import importlib.util
import io
import json
from pathlib import Path
import shutil
import tarfile
import tempfile
import unittest
from unittest import mock

import yaml

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location('native_packages', HERE / 'build-native-packages.py')
BUILDER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(BUILDER)


def metadata(lock):
    return [{'Id': i['image_id'], 'Os': 'linux', 'Architecture': 'amd64',
             'RepoTags': [i['reference'], i['source_reference']]} for i in lock['images']]


def rendering(lock):
    refs = {i['role']: i['reference'] for i in lock['images']}
    compose = {'services': {name: {'image': refs[role]} for name, role in BUILDER.SERVICE_ROLES.items()}}
    docs = [{'kind': 'OBCluster', 'spec': {'observer': {'image': refs['observer']}}},
            {'kind': 'OBLogServiceCluster', 'spec': {'logService': {'image': refs['logservice']}}}]
    containers = [{'name': name, 'image': refs[role]} for name, role in BUILDER.CONTAINER_ROLES.items()]
    docs.append({'kind': 'Deployment', 'spec': {'template': {'spec': {'containers': containers}}}})
    docs.append({'kind': 'ConfigMap', 'data': {'config.yaml': yaml.safe_dump({
        'inspection': {'obdiag': {'image': refs['obdiag']}, 'oceanbase-helper': {'image': refs['oceanbase-helper']}},
        'sqlAnalyzer': {'image': lock['excluded_optional']['sql-analyzer']['reference']}})}})
    return compose, docs


class ReleaseContractTests(unittest.TestCase):
    def setUp(self):
        self.lock = BUILDER.load_lock(HERE / 'release-images.lock.json')
        self.compose, self.helm = rendering(self.lock)

    def test_shared_roles_have_one_actual_identity(self):
        roles = BUILDER.validate_lock(self.lock)
        for role in BUILDER.SHARED:
            self.assertEqual(set(roles[role]['channels']), {'compose', 'helm'})
        self.assertEqual(roles['observer']['reference'], 'oceanbase/oceanbase-ai-cloud-native:4.6.2.1-100000092026090921.al8')
        BUILDER.validate_rendered(self.lock, self.compose, self.helm)
        result = BUILDER.verify_metadata(self.lock, metadata(self.lock))
        self.assertFalse(result['archive_config_bytes_verified'])
        self.assertFalse(result['registry_pull_verified'])

    def test_lock_role_swap_is_rejected(self):
        a, b = self.lock['images'][:2]
        a['reference'], b['reference'] = b['reference'], a['reference']
        a['image_id'], b['image_id'] = b['image_id'], a['image_id']
        with self.assertRaisesRegex(ValueError, 'repository does not match image role'):
            BUILDER.validate_lock(self.lock)

    def test_zone_observer_override_cannot_hide_a_different_component(self):
        self.helm[0]['spec']['topology'] = [{'zone': 'zone1', 'observer': {
            'image': self.helm[1]['spec']['logService']['image']}}]
        with self.assertRaisesRegex(ValueError, 'helm role observer image mismatch'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)

    def test_missing_bootstrap_fails_before_build(self):
        self.lock['images'] = [i for i in self.lock['images'] if i['role'] != 'bootstrap']
        with self.assertRaisesRegex(ValueError, 'Missing required.*bootstrap'):
            BUILDER.validate_lock(self.lock)

    def test_duplicate_role_even_with_distinct_reference_is_rejected(self):
        duplicate = copy.deepcopy(self.lock['images'][0])
        duplicate['reference'] = 'oceanbase/observer:old'
        self.lock['images'].append(duplicate)
        with self.assertRaisesRegex(ValueError, 'duplicate image role'):
            BUILDER.validate_lock(self.lock)

    def test_same_tag_wrong_image_id_is_rejected(self):
        records = metadata(self.lock)
        records[0]['Id'] = 'sha256:' + '0' * 64
        with self.assertRaisesRegex(ValueError, 'identity mismatch'):
            BUILDER.verify_metadata(self.lock, records)

    def test_manifest_digest_is_separate_from_config_id(self):
        item = self.lock['images'][0]
        item['manifest_digest'] = 'sha256:' + 'f' * 64
        records = metadata(self.lock)
        with self.assertRaisesRegex(ValueError, 'manifest digest is not proven'):
            BUILDER.verify_metadata(self.lock, records)
        records[0]['RepoDigests'] = [item['reference'].rsplit(':', 1)[0] + '@' + item['manifest_digest']]
        BUILDER.verify_metadata(self.lock, records)

    def test_conflicting_tag_records_are_rejected(self):
        records = metadata(self.lock)
        duplicate = dict(records[0], Id='sha256:' + '0' * 64)
        records.append(duplicate)
        with self.assertRaisesRegex(ValueError, 'Conflicting identity'):
            BUILDER.verify_metadata(self.lock, records)

    def test_source_alias_alone_does_not_prove_canonical_tag_is_available(self):
        records = metadata(self.lock)
        for record in records:
            record['RepoTags'] = [tag for tag in record['RepoTags'] if tag.startswith('oceanbase-qa/')]
        with self.assertRaisesRegex(ValueError, 'identity mismatch'):
            BUILDER.verify_metadata(self.lock, records)

    def test_wrong_architecture_is_rejected(self):
        records = metadata(self.lock)
        records[0]['Architecture'] = 'arm64'
        with self.assertRaisesRegex(ValueError, 'identity mismatch'):
            BUILDER.verify_metadata(self.lock, records)

    def test_compose_role_swap_fails_even_when_image_set_is_unchanged(self):
        services = self.compose['services']
        services['minio']['image'], services['minio-init']['image'] = services['minio-init']['image'], services['minio']['image']
        with self.assertRaisesRegex(ValueError, 'compose role minio image mismatch'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)

    def test_helm_role_swap_fails_even_when_image_set_is_unchanged(self):
        a, b = self.helm[:2]
        a['spec']['observer']['image'], b['spec']['logService']['image'] = b['spec']['logService']['image'], a['spec']['observer']['image']
        with self.assertRaisesRegex(ValueError, 'helm role observer image mismatch'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)

    def test_previous_native_helm_compose_observer_fork_is_rejected(self):
        # Exact formerly delivered Helm override versus the Compose 4.6.2.1 image.
        self.helm[0]['spec']['observer']['image'] = 'oceanbase-qa/observer:4.6.2.0-tar1.30-20261021'
        with self.assertRaisesRegex(ValueError, 'helm role observer image mismatch'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)

    def test_previous_logservice_alias_is_rejected_even_with_matching_bytes(self):
        self.helm[1]['spec']['logService']['image'] = 'oceanbase-qa/oblogservice:1.3.0-e747a05a-20261019'
        with self.assertRaisesRegex(ValueError, 'helm role logservice image mismatch'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)

    def test_optional_images_cannot_hide_live_pod_or_monitor_dependencies(self):
        self.helm[0]['spec']['monitor'] = {'image': 'oceanbase/obagent:4.2.4'}
        with self.assertRaisesRegex(ValueError, 'optional obagent'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)
        del self.helm[0]['spec']['monitor']
        self.helm.append({'kind': 'Pod', 'spec': {'containers': [
            {'name': 'sql-analyzer', 'image': self.lock['excluded_optional']['sql-analyzer']['reference']}]}})
        with self.assertRaisesRegex(ValueError, 'Unclassified runtime container'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)

    def test_qa_alias_cannot_become_release_reference(self):
        self.lock['images'][0]['reference'] = 'oceanbase-qa/observer:4.6.2.0'
        with self.assertRaisesRegex(ValueError, 'QA source aliases'):
            BUILDER.validate_lock(self.lock)

    def test_removed_runtime_component_is_not_silently_accepted(self):
        self.helm = [d for d in self.helm if d['kind'] != 'ConfigMap']
        with self.assertRaisesRegex(ValueError, 'missing required image roles'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)

    def test_archive_timestamp_rejects_epoch_future_and_invalid_input(self):
        for value in (0, '-1', 'not-an-epoch', '3.5', '99999999999999'):
            with self.subTest(value=value), self.assertRaises(ValueError):
                BUILDER.release_timestamp(value)
        self.assertEqual(BUILDER.release_timestamp('946684800'), 946684800)

    def test_tar_headers_use_a_stable_reasonable_date(self):
        with tempfile.TemporaryDirectory() as root:
            root = Path(root)
            files = root / 'bundle'
            files.mkdir()
            (files / 'file.txt').write_text('release fixture')
            target = root / 'bundle.tar.gz'
            BUILDER.tar_tree(files, target)
            with tarfile.open(target, 'r:gz') as archive:
                self.assertEqual({item.mtime for item in archive}, {946684800})

    def test_archive_config_hash_is_computed_from_bytes(self):
        with tempfile.TemporaryDirectory() as root:
            target = Path(root) / 'images.tar.gz'
            config = json.dumps({'architecture': 'amd64', 'os': 'linux', 'config': {'Labels': {'fixture': 'yes'}}}).encode()
            expected = 'sha256:' + hashlib.sha256(config).hexdigest()
            with tarfile.open(target, 'w:gz') as archive:
                for name, data in [('untrusted-filename.json', config), ('manifest.json', json.dumps([
                    {'Config': 'untrusted-filename.json', 'RepoTags': ['fixture:1'], 'Layers': []}]).encode())]:
                    info = tarfile.TarInfo(name); info.size = len(data)
                    archive.addfile(info, io.BytesIO(data))
            result = BUILDER.archive_metadata(target)
            self.assertEqual(result[0]['Id'], expected)
            self.assertEqual(result[0]['RepoTags'], ['fixture:1'])


@unittest.skipUnless(shutil.which('helm') and shutil.which('docker'), 'Helm and Docker Compose CLIs required')
class RealPackagingTests(unittest.TestCase):
    def test_both_deliveries_are_rendered_and_share_exact_refs(self):
        with tempfile.TemporaryDirectory() as root:
            root = Path(root)
            lock = BUILDER.load_lock(HERE / 'release-images.lock.json')
            evidence = root / 'synthetic-inspect-fixture.json'
            evidence.write_text(json.dumps(metadata(lock)))
            args = argparse.Namespace(source=HERE.parent.parent, lock=HERE / 'release-images.lock.json',
                                      output=root / 'delivery', version='test', docker_metadata=evidence,
                                      image_archive=None, helm='helm', docker='docker')
            # A poisoned shell must not override the generated release environment.
            with mock.patch.dict('os.environ', {'OBSERVER_IMAGE': 'bad.example/observer:old',
                                               'IMAGE_REGISTRY': 'bad.example', 'SOURCE_DATE_EPOCH': '1704067200'}):
                result = BUILDER.build(args)
            self.assertEqual(result['identity_evidence']['source'], 'supplied-metadata-snapshot')
            self.assertFalse(result['identity_evidence']['live_docker_checked'])
            self.assertFalse(result['live_installation_tested'])
            self.assertEqual(len(result['artifacts']), 2)
            self.assertEqual(result['archive_timestamp'], 1704067200)
            for profile, rendered in result['rendered_roles'].items():
                for role in BUILDER.SHARED:
                    self.assertEqual(rendered['compose'][role], rendered['helm'][role], profile)
            compose = args.output / 'oceanbase-lakehouse-compose-test'
            helm = args.output / 'oceanbase-lakehouse-native-helm-test'
            c = json.loads((compose / 'image-lock.json').read_text())
            h = json.loads((helm / 'image-lock.json').read_text())
            self.assertEqual({i['role'] for i in c['images']}, BUILDER.COMPOSE_ROLES)
            self.assertEqual({i['role'] for i in h['images']}, BUILDER.HELM_ROLES)
            self.assertEqual((compose / 'release-images.lock.json').read_bytes(), (helm / 'release-images.lock.json').read_bytes())
            with tarfile.open(next(helm.glob('oceanbase-lakehouse-*.tgz')), 'r:gz') as chart:
                self.assertFalse(any(p.name.endswith('install.sh') or p.name.endswith('_test.go') for p in chart))
                values = yaml.safe_load(chart.extractfile('oceanbase-lakehouse/values.yaml'))
                refs = {i['role']: i['reference'] for i in lock['images']}
                self.assertEqual(values['observer']['image'], refs['observer'])
                self.assertEqual(values['logService']['image'], refs['logservice'])
            # Re-read the delivered archive, including its hidden Compose .env.
            with tarfile.open(args.output / 'oceanbase-lakehouse-compose-test.tar.gz', 'r:gz') as archive:
                self.assertIsNotNone(archive.getmember('oceanbase-lakehouse-compose-test/.env'))
                self.assertEqual({item.mtime for item in archive}, {1704067200})


if __name__ == '__main__':
    unittest.main()

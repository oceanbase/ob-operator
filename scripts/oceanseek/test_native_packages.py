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
import re
from pathlib import Path
import shlex
import shutil
import subprocess
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
    records = []
    for item in lock['images']:
        record = {'Id': item['image_id'], 'Os': 'linux', 'Architecture': 'amd64',
                  'RepoTags': [item['reference'], item['source_reference']]}
        if item.get('manifest_digest'):
            repository = item['reference'].rsplit(':', 1)[0]
            record['RepoDigests'] = [repository + '@' + item['manifest_digest']]
        records.append(record)
    return records


def rendering(lock):
    refs = {i['role']: i['reference'] for i in lock['images']}
    compose = {'services': {name: {'image': refs[role]} for name, role in BUILDER.SERVICE_ROLES.items()}}
    docs = [{'kind': 'OBCluster', 'spec': {'observer': {'image': refs['observer']}, 'monitor': {'image': refs['obagent']}}},
            {'kind': 'OBLogServiceCluster', 'spec': {'logService': {'image': refs['logservice']}}}]
    containers = [{'name': name, 'image': refs[role]} for name, role in BUILDER.CONTAINER_ROLES.items()]
    docs.append({'kind': 'Deployment', 'spec': {'template': {'spec': {'containers': containers}}}})
    docs.append({'kind': 'ConfigMap', 'data': {'config.yaml': yaml.safe_dump({
        'inspection': {'obdiag': {'image': refs['obdiag']}, 'oceanbase-helper': {'image': refs['oceanbase-helper']}},
        'sqlAnalyzer': {'image': lock['excluded_optional']['sql-analyzer']['reference']}})}})
    return compose, docs


class HelmSourceProvenanceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        spec = importlib.util.spec_from_file_location('helm_release', HERE / 'build-helm-release.py')
        cls.builder = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.builder)

    def test_exported_source_without_git_executable(self):
        with tempfile.TemporaryDirectory() as source, mock.patch.dict('os.environ', {'PATH': source}):
            metadata, patch = self.builder.source_provenance(Path(source), {})
        self.assertEqual(metadata, {'source_commit': None, 'source_worktree_dirty': None,
                                    'git_metadata_available': False})
        self.assertEqual(patch, '')

    @unittest.skipUnless(shutil.which('helm'), 'Helm CLI required')
    def test_packaged_chart_preserves_cert_manager_license_and_notices(self):
        prerequisites = HERE / 'prerequisites'
        manifest = (prerequisites / 'cert-manager-locked.yaml').read_text()
        lock = BUILDER.load_lock(HERE / 'testdata/native-images.lock.json')
        for image in lock['images']:
            image['manifest_digest'] = image['image_id']  # synthetic packaging fixture
        for ref, role, digest in sorted(set(re.findall(
                r'([^"\s=]+/(cert-manager-[a-z0-9]+)@(sha256:[0-9a-f]{64}))', manifest))):
            lock['images'].append({'role': role, 'reference': ref, 'manifest_digest': digest,
                                   'image_id': digest, 'platform': lock['platform']})
        with tempfile.TemporaryDirectory() as tmp:
            result = self.builder.build_release(HERE.parents[1], lock, Path(tmp) / 'release')
            with tarfile.open(result['chart']) as chart:
                prefix = 'oceanseek/licenses/cert-manager/'
                license_text = chart.extractfile(prefix + 'LICENSE').read()
                self.assertEqual(hashlib.sha256(license_text).hexdigest(),
                                 'c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4')
                notice = chart.extractfile(prefix + 'NOTICE.txt').read().decode()
                self.assertIn('Copyright 2022 The cert-manager Authors.', notice)
                self.assertIn('Apache License, Version 2.0', notice)
                foundation = json.load(chart.extractfile('oceanseek/files/prerequisites/foundation.json'))
                self.assertIn('licenses/cert-manager/NOTICE.txt', foundation['certManagerNotice'])
                crds = [name for name in chart.getnames()
                        if name.startswith('oceanseek/crds/') and name.endswith('.cert-manager.io.yaml')]
                self.assertTrue(crds)
                for name in crds:
                    self.assertIn(b'licenses/cert-manager/NOTICE.txt', chart.extractfile(name).readline())


class ReleaseContractTests(unittest.TestCase):
    def setUp(self):
        self.lock = BUILDER.load_lock(HERE / 'testdata/native-images.lock.json')
        self.compose, self.helm = rendering(self.lock)

    def test_obproxy_is_locked_for_helm_only(self):
        roles = BUILDER.validate_lock(self.lock)
        self.assertEqual(roles['obproxy']['channels'], ['helm'])
        values = BUILDER.image_values(self.lock)
        self.assertEqual(values['obproxy']['image'], roles['obproxy']['reference'])
        self.assertNotIn('obproxy', BUILDER.COMPOSE_ROLES)
        missing = copy.deepcopy(self.lock)
        missing['images'] = [item for item in missing['images'] if item['role'] != 'obproxy']
        with self.assertRaisesRegex(ValueError, 'Missing required.*obproxy'):
            BUILDER.validate_lock(missing)

    def test_shared_roles_have_one_actual_identity(self):
        roles = BUILDER.validate_lock(self.lock)
        for role in BUILDER.SHARED:
            self.assertEqual(set(roles[role]['channels']), {'compose', 'helm'})
        self.assertEqual(roles['observer']['reference'], 'registry.example.invalid/test/oceanseek:1.0.0')
        BUILDER.validate_rendered(self.lock, self.compose, self.helm)
        result = BUILDER.verify_metadata(self.lock, metadata(self.lock))
        self.assertFalse(result['archive_config_bytes_verified'])
        self.assertFalse(result['registry_pull_verified'])

    def test_separate_arm_readiness_is_used_by_both_helm_consumers(self):
        self.lock['platform'] = 'linux/arm64'
        for image in self.lock['images']:
            image['platform'] = self.lock['platform']
        ref = 'example.invalid/oceanseek-readiness:1.0.0'
        self.lock['images'].append({
            'role': 'readiness', 'reference': ref, 'image_id': 'sha256:' + 'a' * 64,
            'platform': self.lock['platform'], 'channels': ['helm'],
        })
        self.assertEqual(BUILDER.image_values(self.lock)['native']['readiness']['image'], ref)
        for doc in self.helm:
            if doc['kind'] == 'Deployment':
                for container in doc['spec']['template']['spec']['containers']:
                    if container['name'] in ('discover-observers', 'readiness'):
                        container['image'] = ref
        roles = BUILDER.validate_rendered(self.lock, self.compose, self.helm)
        self.assertEqual(roles['helm']['readiness'], ref)

    def test_bootstrap_release_tag_and_manual_inputs(self):
        path = HERE.parents[1] / '.github/workflows/oceanseek-compose.yml'
        workflow = yaml.load(path.read_text(), Loader=yaml.BaseLoader)
        push = workflow['on']['push']
        self.assertEqual(push['tags'], ['oceanseek-bootstrap-[0-9]+.[0-9]+.[0-9]+'])
        self.assertIn('**', push['branches'])
        job = workflow['jobs']['bootstrap-image']
        self.assertIn("startsWith(github.ref, 'refs/tags/oceanseek-bootstrap-')", job['if'])
        resolve = next(step['run'] for step in job['steps'] if step.get('id') == 'release')
        for event, ref, version, manual_publish, result in (
            ('push', 'oceanseek-bootstrap-1.2.3', '', '', ('1.2.3', 'true')),
            ('push', 'oceanseek-bootstrap-1.2.3-dev', '', '', None),
            ('workflow_dispatch', 'codex/oceanseek-compose', '2.3.4', 'false', ('2.3.4', 'false')),
            ('workflow_dispatch', 'codex/oceanseek-compose', '2.3.4', 'true', ('2.3.4', 'true')),
            ('workflow_dispatch', 'codex/oceanseek-compose', 'latest', 'true', None),
        ):
            with self.subTest(event=event, ref=ref, version=version), tempfile.TemporaryDirectory() as tmp:
                output = Path(tmp) / 'outputs'
                proc = subprocess.run(['bash', '-e', '-c', resolve], env={
                    'GITHUB_EVENT_NAME': event, 'GITHUB_REF_NAME': ref,
                    'DISPATCH_VERSION': version, 'DISPATCH_PUBLISH': manual_publish,
                    'GITHUB_OUTPUT': str(output),
                })
                if result is None:
                    self.assertNotEqual(proc.returncode, 0)
                    self.assertFalse(output.exists())
                else:
                    self.assertEqual(proc.returncode, 0)
                    self.assertEqual(output.read_text().splitlines(),
                                     ['version=' + result[0], 'publish=' + result[1]])

    def test_public_source_values_select_independent_versions(self):
        for registry in ('quay.io', 'ghcr.io', 'docker.io'):
            args = ['python3', str(HERE / 'public-image-values.py'), '--registry', registry,
                    '--component-namespace', 'example', '--operator-version', '1.2.3',
                    '--dashboard-version', '4.5.6', '--helper-version', '7.8.9']
            values = json.loads(subprocess.check_output(args))
            self.assertEqual(values['native']['readiness']['image'], registry + '/example/ob-operator:1.2.3')
            self.assertEqual(values['dashboard']['image'], registry + '/example/oceanbase-dashboard:4.5.6')
            self.assertEqual(values['prerequisites']['image'], registry + '/example/oceanbase-helper:7.8.9')
            self.assertEqual(values['dashboard']['inspection']['oceanbaseHelper']['image'], values['prerequisites']['image'])
            self.assertEqual(values['observer']['image'], registry + '/oceanbase/oceanseek:4.6.3.0-482026112114')

    def test_lock_role_swap_is_rejected(self):
        a, b = self.lock['images'][:2]
        a['reference'], b['reference'] = b['reference'], a['reference']
        a['image_id'], b['image_id'] = b['image_id'], a['image_id']
        with self.assertRaisesRegex(ValueError, 'repository does not match image role'):
            BUILDER.validate_lock(self.lock)

    def test_single_repository_distribution_keeps_role_identity(self):
        self.lock['distribution_repository'] = 'registry.example/team/lakebase'
        for item in self.lock['images']:
            original = item['reference']
            item['upstream_reference'] = original
            item['reference'] = self.lock['distribution_repository'] + ':' + original.rsplit('/', 1)[-1].replace(':', '-', 1)
        BUILDER.validate_lock(self.lock)
        compose, helm = rendering(self.lock)
        BUILDER.validate_rendered(self.lock, compose, helm)
        self.lock['images'][0]['reference'] = self.lock['images'][1]['reference']
        with self.assertRaisesRegex(ValueError, 'Distribution reference does not match image role'):
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
        records[0].pop('RepoDigests', None)
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
        with self.assertRaisesRegex(ValueError, 'helm role obagent image mismatch'):
            BUILDER.validate_rendered(self.lock, self.compose, self.helm)
        self.helm[0]['spec']['monitor']['image'] = BUILDER.validate_lock(self.lock)['obagent']['reference']
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

    def test_delivery_projection_does_not_disclose_builder_metadata(self):
        lock = copy.deepcopy(self.lock)
        lock['build_notes'] = 'private-marker'
        for image in lock['images']:
            image['source_reference'] = 'private-marker'
            image['diagnostics'] = {'password': 'private-marker'}
        projected = BUILDER.delivery_lock(lock)
        self.assertNotIn('private-marker', json.dumps(projected))
        self.assertEqual(BUILDER.image_values(projected), BUILDER.image_values(lock))
        for old, new in zip(lock['images'], projected['images']):
            self.assertEqual(old['image_id'], new['image_id'])
            self.assertEqual(old.get('manifest_digest'), new.get('manifest_digest'))
        with self.assertRaisesRegex(ValueError, 'public registry'):
            BUILDER.check_public_references(projected)
        for image in projected['images']:
            image['reference'] = 'quay.io/oceanbase/' + image['reference'].rsplit('/', 1)[-1]
        BUILDER.check_public_references(projected)

    def test_chart_snapshot_excludes_local_private_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            source, target = Path(tmp) / 'source', Path(tmp) / 'target'
            source.mkdir()
            (source / 'Chart.yaml').write_text('name: example')
            (source / '.env.local').write_text('private-marker')
            (source / '.oceanseek-install').mkdir()
            (source / '.oceanseek-install' / 'values.yaml').write_text('private-marker')
            (source / 'local-values.yaml').symlink_to(source / '.env.local')
            BUILDER.copy_chart(source, target)
            self.assertEqual([p.name for p in target.iterdir()], ['Chart.yaml'])

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

    def test_arm64_compose_delivery_uses_locked_platform(self):
        self.lock['platform'] = 'linux/arm64'
        for image in self.lock['images']:
            image['platform'] = 'linux/arm64'
        self.assertIn('OCEANSEEK_PLATFORM=linux/arm64', BUILDER.compose_env(self.lock).splitlines())
        with tempfile.TemporaryDirectory() as root:
            target = Path(root) / 'compose.yaml'
            BUILDER.write_standalone_compose(HERE.parents[1], target, self.lock)
            services = yaml.safe_load(target.read_text())['services']
            self.assertEqual({s['platform'] for s in services.values()}, {'linux/arm64'})

    def test_standalone_compose_avoids_swarm_only_config_modes(self):
        with tempfile.TemporaryDirectory() as root:
            target = Path(root) / 'compose.yaml'
            BUILDER.write_standalone_compose(HERE.parents[1], target, self.lock)
            rendered = yaml.safe_load(target.read_text())
            for service in rendered['services'].values():
                for config in service.get('configs', []):
                    self.assertFalse({'uid', 'gid', 'mode'} & config.keys())
                entrypoint = service.get('entrypoint') or []
                if any(str(arg).startswith('/opt/oceanseek/') for arg in entrypoint):
                    self.assertIn(entrypoint[0], ['sh', 'bash'])
            self.assertEqual(rendered['services']['observer']['entrypoint'][0], 'bash')
            self.assertIn('${COMPOSE_PROJECT_NAME:-oceanseek}', rendered['name'])

    def test_standalone_observer_entrypoint_runs_nonexecutable_monitoring_helper(self):
        with tempfile.TemporaryDirectory() as root:
            root = Path(root)
            target = root / 'compose.yaml'
            BUILDER.write_standalone_compose(HERE.parents[1], target, self.lock)
            rendered = yaml.safe_load(target.read_text())
            service = rendered['services']['observer']
            content = rendered['configs']['scripts-observer-entrypoint-sh']['content']
            calls = [line.strip() for line in content.splitlines()
                     if 'obshell-monitoring-register.sh' in line and line.strip().endswith('&')]
            self.assertEqual(len(calls), 1, calls)
            helper = root / 'obshell-monitoring-register.sh'
            helper.write_text('#!/usr/bin/env bash\nprintf "monitoring helper started\\n"\n')
            helper.chmod(0o444)
            self.assertEqual(helper.stat().st_mode & 0o777, 0o444)
            test_script = calls[0].replace('/opt/oceanseek/obshell-monitoring-register.sh', shlex.quote(str(helper))) + '\nwait $!\n'
            completed = subprocess.run(
                ['bash', '-c', test_script], cwd=root, text=True,
                capture_output=True, timeout=5,
            )
            self.assertEqual(completed.returncode, 0, completed.stdout + completed.stderr)
            self.assertIn('monitoring helper started', completed.stdout)
            self.assertIn('/opt/oceanseek/observer-entrypoint.sh', service['entrypoint'])

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
    def test_single_yaml_needs_no_local_scripts_or_env(self):
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory) / 'compose.yaml'
            lock = BUILDER.load_lock(HERE / 'testdata/native-images.lock.json')
            BUILDER.write_standalone_compose(HERE.parents[1], target, lock)
            rendered = json.loads(BUILDER.run(['docker', 'compose', '--env-file', '/dev/null',
                                              '-f', target, 'config', '--format', 'json'],
                                             cwd=target.parent, env=BUILDER.clean_compose_env()))
            self.assertEqual(list(target.parent.iterdir()), [target])
            for service in rendered['services'].values():
                self.assertFalse(any(mount['type'] == 'bind' for mount in service.get('volumes', [])))
                for config in service.get('configs', []):
                    self.assertFalse({'uid', 'gid', 'mode'} & config.keys())
                entrypoint = service.get('entrypoint') or []
                if any(str(arg).startswith('/opt/oceanseek/') for arg in entrypoint):
                    self.assertIn(entrypoint[0], ['sh', 'bash'])
            content = rendered['configs']['scripts-observer-entrypoint-sh']['content']
            # Compose config serializes literal dollars escaped again so its output
            # can itself be consumed by Compose. Runtime mounting is tested separately.
            original = (HERE.parents[1] / 'deploy/compose/oceanseek/scripts/observer-entrypoint.sh').read_text()
            self.assertEqual(content, original.replace('$', '$$'))
            configs = rendered['services']['observer']['configs']
            self.assertTrue(any(item['target'] == '/opt/oceanseek/observer-entrypoint.sh' for item in configs))

    def test_standalone_digest_lock_rejects_missing_or_mutable_images(self):
        original = BUILDER.load_lock(HERE / 'testdata/native-images.lock.json')
        lock = {'schemaVersion': 1, 'platform': 'linux/amd64', 'delivery_channel': 'compose',
                'images': [dict(item, reference='registry.example/image@'+item['image_id'])
                           for item in original['images'] if item['role'] in BUILDER.COMPOSE_ROLES]}
        with tempfile.TemporaryDirectory() as root:
            target = Path(root) / 'compose.yaml'
            BUILDER.write_standalone_compose(HERE.parents[1], target, lock)
            rendered = yaml.safe_load(target.read_text())
            self.assertTrue(all('@sha256:' in v['image'] for v in rendered['services'].values()))
            missing = copy.deepcopy(lock)
            missing['images'].pop()
            with self.assertRaisesRegex(ValueError, 'Incomplete'):
                BUILDER.write_standalone_compose(HERE.parents[1], target, missing)
            lock['images'][0]['reference'] = 'registry.example/image:latest'
            with self.assertRaisesRegex(ValueError, 'Invalid digest'):
                BUILDER.write_standalone_compose(HERE.parents[1], target, lock)

    def test_single_repository_charts_and_publication_evidence(self):
        with tempfile.TemporaryDirectory() as root:
            root = Path(root)
            lock = BUILDER.load_lock(HERE / 'testdata/native-images.lock.json')
            lock['distribution_repository'] = 'registry.example/team/lakebase'
            proof = {}
            for item in lock['images']:
                original = item['reference']
                item['upstream_reference'] = original
                item['reference'] = lock['distribution_repository'] + ':' + original.rsplit('/', 1)[-1].replace(':', '-', 1)
                item['manifest_digest'] = 'sha256:' + hashlib.sha256(item['reference'].encode()).hexdigest()
                item['publication'] = 'verified-registry-pull'
                proof[item['role']] = dict(item, registry_pull_verified=True)
            lock_file = root / 'fixture-lock.json'
            lock_file.write_text(json.dumps(lock))
            evidence = root / 'synthetic-inspect-fixture.json'
            evidence.write_text(json.dumps(metadata(lock)))
            publication = root / 'synthetic-publication-fixture.json'
            publication.write_text(json.dumps(proof))
            args = argparse.Namespace(source=HERE.parent.parent, lock=lock_file, output=root / 'delivery',
                                      version='fixture', docker_metadata=evidence, image_archive=None,
                                      publication_report=publication, helm='helm', docker='docker')
            report = BUILDER.build(args)
            self.assertTrue(report['registry_publication_verified'])
            for profile in report['rendered_roles'].values():
                for roles in profile.values():
                    self.assertTrue(all(ref.startswith(lock['distribution_repository'] + ':') for ref in roles.values()))
            helm = args.output / 'oceanseek-native-helm-fixture'
            with tarfile.open(next(helm.glob('ob-operator-*.tgz'))) as archive:
                values = yaml.safe_load(archive.extractfile('ob-operator/values.yaml'))
                self.assertEqual(values['kubeRbacProxyImage'], proof['kube-rbac-proxy']['reference'])
            self.assertNotIn('尚未验证仓库发布', (helm / 'OFFLINE.md').read_text())
            proof['observer']['image_id'] = 'sha256:' + '0' * 64
            publication.write_text(json.dumps(proof))
            args.output = root / 'must-not-build'
            with self.assertRaisesRegex(ValueError, 'publication evidence mismatch'):
                BUILDER.build(args)
            self.assertFalse(args.output.exists())

    def test_both_deliveries_are_rendered_and_share_exact_refs(self):
        with tempfile.TemporaryDirectory() as root:
            root = Path(root)
            lock = BUILDER.load_lock(HERE / 'testdata/native-images.lock.json')
            evidence = root / 'synthetic-inspect-fixture.json'
            evidence.write_text(json.dumps(metadata(lock)))
            args = argparse.Namespace(source=HERE.parent.parent, lock=HERE / 'testdata/native-images.lock.json',
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
            compose = args.output / 'oceanseek-compose-test'
            helm = args.output / 'oceanseek-native-helm-test'
            self.assertFalse((compose / 'compose.aistor.yaml').exists())
            compose_env = (compose / '.env').read_text()
            self.assertIn('HOST_ACCESS_IP=', compose_env)
            self.assertIn('HOST_BIND_ADDRESS=0.0.0.0', compose_env)
            self.assertIn('AUTO_REGISTER_MONITORING=true', compose_env)
            self.assertIn('PROMETHEUS_INTERNAL_URL=http://prometheus:9090/', compose_env)
            self.assertIn('ALERTMANAGER_INTERNAL_URL=http://alertmanager:9093', compose_env)
            self.assertIn('OBSERVER_SQL_BIND_ADDRESS=0.0.0.0', compose_env)
            self.assertIn('OBSHELL_BIND_ADDRESS=0.0.0.0', compose_env)
            self.assertIn('OBSHELL_HOST_PORT=32886', compose_env)
            self.assertIn('${IMAGE_REGISTRY:-registry.example.invalid/test}/minio:', compose_env)
            self.assertIn('${IMAGE_REGISTRY:-registry.example.invalid/test}/mc:', compose_env)
            readme = (compose / 'README.md').read_text()
            self.assertIn('docker compose up -d && docker compose logs -f observer-bootstrap', readme)
            c = json.loads((compose / 'image-lock.json').read_text())
            h = json.loads((helm / 'image-lock.json').read_text())
            self.assertEqual({i['role'] for i in c['images']}, BUILDER.COMPOSE_ROLES)
            self.assertEqual({i['role'] for i in h['images']}, BUILDER.HELM_ROLES)
            self.assertEqual((compose / 'release-images.lock.json').read_bytes(), (helm / 'release-images.lock.json').read_bytes())
            with tarfile.open(next(helm.glob('oceanseek-*.tgz')), 'r:gz') as chart:
                self.assertFalse(any(p.name.endswith('install.sh') or p.name.endswith('_test.go') for p in chart))
                values = yaml.safe_load(chart.extractfile('oceanseek/values.yaml'))
                refs = {i['role']: i['reference'] for i in lock['images']}
                self.assertEqual(values['observer']['image'], refs['observer'])
                self.assertEqual(values['logService']['image'], refs['logservice'])
                self.assertTrue(values['dashboard']['enabled'])
                self.assertTrue(values['native']['enabled'])
                self.assertEqual(values['logService']['resource']['cpu'], '2')
            quickstart = (helm / 'QUICKSTART.md').read_text()
            install = next(line for line in quickstart.splitlines() if line.startswith('helm upgrade'))
            self.assertIn(next(helm.glob('oceanseek-*.tgz')).name, install)
            self.assertIn('--wait --wait-for-jobs --timeout 35m', install)
            self.assertNotIn(' -f ', install)
            for directory in (compose, helm):
                mirror_root = directory / 'registry-mirrors'
                self.assertTrue((mirror_root / 'README.md').is_file())
                self.assertTrue((mirror_root / 'configure-docker.py').is_file())
                docker_config = json.loads((mirror_root / 'docker-daemon.json').read_text())
                self.assertEqual(docker_config['registry-mirrors'], [])
                self.assertEqual(docker_config['max-concurrent-downloads'], 10)
                self.assertEqual(docker_config['max-concurrent-uploads'], 5)
                registry_config = yaml.safe_load((mirror_root / 'k3s-registries.yaml').read_text())
                self.assertEqual(registry_config['mirrors'], {})
                checksums = (directory / 'SHA256SUMS').read_text().splitlines()
                for path in mirror_root.rglob('*'):
                    self.assertNotIn(path.name, ('tests', '__pycache__'))
                    if path.is_file():
                        self.assertIn(BUILDER.sha256(path) + '  ' + path.relative_to(directory).as_posix(), checksums)
                with tarfile.open(args.output / (directory.name + '.tar.gz'), 'r:gz') as archive:
                    prefix = directory.name + '/registry-mirrors/'
                    for name in ('README.md', 'configure-docker.py', 'docker-daemon.json', 'k3s-registries.yaml'):
                        self.assertEqual(archive.extractfile(prefix + name).read(), (mirror_root / name).read_bytes())
            # Re-read the delivered archive, including its hidden Compose .env.
            with tarfile.open(args.output / 'oceanseek-compose-test.tar.gz', 'r:gz') as archive:
                self.assertIsNotNone(archive.getmember('oceanseek-compose-test/.env'))
                self.assertEqual({item.mtime for item in archive}, {1704067200})


if __name__ == '__main__':
    unittest.main()

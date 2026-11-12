#!/usr/bin/env python3
"""Build matching native Helm and Compose bundles from one reviewed image lock.

Build dependencies: Python 3.8+, PyYAML, Helm and Docker Compose. No Kubernetes
changes or registry publication are performed. Images are not included. Identity
validation uses live Docker inspect, supplied inspect metadata, or actual Docker
save archives; the validation report explicitly records which evidence was used.
"""
import argparse
import copy
import gzip
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tarfile
import tempfile
import time

import yaml

HERE = Path(__file__).resolve().parent
SHARED = {'observer', 'logservice', 'minio', 'mc', 'prometheus', 'alertmanager'}
COMPOSE_ROLES = SHARED | {'bootstrap', 'obagent'}
HELM_ROLES = SHARED | {'operator', 'dashboard', 'obdiag', 'oceanbase-helper',
                       'utility', 'prometheus', 'alertmanager', 'kube-rbac-proxy'}
CHANNELS = {'compose': COMPOSE_ROLES, 'helm': HELM_ROLES}
SERVICE_ROLES = {'observer': 'observer', 'minio': 'minio', 'minio-init': 'mc',
                 'credentials-init': 'bootstrap', 'host-info-init': 'bootstrap',
                 'observer-bootstrap': 'bootstrap',
                 'obagent': 'obagent', 'prometheus': 'prometheus',
                 'alertmanager': 'alertmanager',
                 'logservice-bootstrap': 'logservice',
                 **{'logservice-%d' % i: 'logservice' for i in (1, 2, 3)}}
CONTAINER_ROLES = {'manager': 'operator', 'kube-rbac-proxy': 'kube-rbac-proxy',
                   'dashboard': 'dashboard', 'prometheus': 'prometheus',
                   'alertmanager': 'alertmanager', 'minio': 'minio',
                   'initialize': 'mc', 'readiness': 'operator',
                   'prometheus-storage-owner-init': 'utility', 'cleanup': 'utility'}
ENV_ROLES = {'OBSERVER_IMAGE': 'observer', 'LOGSERVICE_IMAGE': 'logservice',
             'BOOTSTRAP_IMAGE': 'bootstrap', 'MINIO_IMAGE': 'minio', 'MINIO_MC_IMAGE': 'mc',
             'OBAGENT_IMAGE': 'obagent', 'PROMETHEUS_IMAGE': 'prometheus',
             'ALERTMANAGER_IMAGE': 'alertmanager'}
ROLE_REPOSITORIES = {'observer': 'oceanbase-ai-cloud-native', 'logservice': 'oblogservice',
                     'minio': 'minio', 'mc': 'mc', 'bootstrap': 'lakehouse-bootstrap',
                     'obagent': 'obagent',
                     'operator': 'ob-operator', 'dashboard': 'oceanbase-dashboard',
                     'obdiag': 'obdiag', 'oceanbase-helper': 'oceanbase-helper',
                     'utility': 'busybox', 'prometheus': 'prometheus',
                     'alertmanager': 'alertmanager', 'kube-rbac-proxy': 'kube-rbac-proxy'}
IMAGE_ID = re.compile(r'^sha256:[0-9a-f]{64}$')


def sha256(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as stream:
        for part in iter(lambda: stream.read(8 * 1024 * 1024), b''):
            h.update(part)
    return h.hexdigest()


def run(argv, cwd=None, env=None):
    result = subprocess.run(list(map(str, argv)), cwd=cwd, env=env, text=True,
                            capture_output=True, timeout=300)
    if result.returncode:
        raise ValueError('Command failed: %s\n%s' % (' '.join(map(str, argv)), result.stderr.strip()))
    return result.stdout


def normalize_ref(ref):
    if ref.startswith('docker.io/'):
        ref = ref[10:]
    return ref[8:] if ref.startswith('library/') else ref


def load_lock(path):
    lock = json.loads(Path(path).read_text())
    validate_lock(lock)
    return lock


def validate_lock(lock):
    if lock.get('schemaVersion') != 1 or lock.get('platform') != 'linux/amd64':
        raise ValueError('Expected schemaVersion 1 and linux/amd64 release platform')
    roles, refs = {}, set()
    for item in lock.get('images', []):
        role, ref = item.get('role'), item.get('reference', '')
        if role in roles or role not in COMPOSE_ROLES | HELM_ROLES:
            raise ValueError('Unexpected or duplicate image role: ' + str(role))
        if (not re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9._:/-]*', ref) or
                ':' not in ref.rsplit('/', 1)[-1] or ref.endswith(':latest')):
            raise ValueError('Canonical reference requires an explicit non-latest tag: ' + ref)
        if 'oceanbase-qa/' in ref:
            raise ValueError('QA source aliases must not be release references: ' + ref)
        distribution_repo = lock.get('distribution_repository')
        if distribution_repo:
            upstream = item.get('upstream_reference', '')
            if (':' not in upstream.rsplit('/', 1)[-1] or
                    upstream.rsplit('/', 1)[-1].rsplit(':', 1)[0] != ROLE_REPOSITORIES[role]):
                raise ValueError('Upstream repository does not match image role: ' + role)
            expected = distribution_repo + ':' + ROLE_REPOSITORIES[role] + '-' + upstream.rsplit(':', 1)[1]
            if ref != expected:
                raise ValueError('Distribution reference does not match image role: ' + role)
        elif ref.rsplit('/', 1)[-1].rsplit(':', 1)[0] != ROLE_REPOSITORIES[role]:
            raise ValueError('Canonical repository does not match image role: ' + role)
        if normalize_ref(ref) in refs:
            raise ValueError('Duplicate canonical reference: ' + ref)
        if not IMAGE_ID.fullmatch(item.get('image_id', '')) or item.get('platform') != lock['platform']:
            raise ValueError('Missing immutable image ID or mismatched platform: ' + ref)
        digest = item.get('manifest_digest')
        if digest is not None and not IMAGE_ID.fullmatch(digest):
            raise ValueError('Invalid registry manifest digest: ' + ref)
        channels = item.get('channels', [])
        expected = {channel for channel, wanted in CHANNELS.items() if role in wanted}
        if len(channels) != len(set(channels)) or set(channels) != expected:
            raise ValueError('Incorrect delivery channels for role: ' + role)
        roles[role] = item
        refs.add(normalize_ref(ref))
    missing = (COMPOSE_ROLES | HELM_ROLES) - roles.keys()
    if missing:
        raise ValueError('Missing required image roles: ' + ', '.join(sorted(missing)))
    return roles


def verify_metadata(lock, metadata):
    """Check supplied Docker inspect records. This does NOT verify archive bytes."""
    roles = validate_lock(lock)
    actual, repo_digests = {}, {}
    if not isinstance(metadata, list):
        raise ValueError('Docker metadata must be a docker image inspect JSON array')
    for record in metadata:
        identity = (record.get('Id'), record.get('Os', '') + '/' + record.get('Architecture', ''))
        if not IMAGE_ID.fullmatch(identity[0] or ''):
            raise ValueError('Docker metadata lacks a full config image ID')
        for ref in record.get('RepoTags') or []:
            ref = normalize_ref(ref)
            if ref in actual and actual[ref] != identity:
                raise ValueError('Conflicting identity for image reference: ' + ref)
            actual[ref] = identity
            repo_digests.setdefault(ref, set()).update(record.get('RepoDigests') or [])
    for item in roles.values():
        ref = normalize_ref(item['reference'])
        expected = (item['image_id'], item['platform'])
        if actual.get(ref) != expected:
            raise ValueError('Image identity mismatch for %s: expected %s, got %s' %
                             (item['reference'], expected, actual.get(ref)))
        source_ref = normalize_ref(item.get('source_reference', item['reference']))
        if source_ref in actual and actual[source_ref] != expected:
            raise ValueError('Source alias identity mismatch: ' + source_ref)
        if item.get('manifest_digest'):
            repository = item['reference'].rsplit(':', 1)[0]
            expected_digest = normalize_ref(repository + '@' + item['manifest_digest'])
            if expected_digest not in {normalize_ref(value) for value in repo_digests[ref]}:
                raise ValueError('Registry manifest digest is not proven by metadata: ' + ref)
    return {'kind': 'docker-inspect-metadata', 'images_checked': len(roles),
            'archive_config_bytes_verified': False, 'image_layers_verified': False,
            'registry_pull_verified': False}


def archive_metadata(path):
    """Read config bytes from a Docker save archive, checking actual SHA-256."""
    configs, manifests = {}, []
    with tarfile.open(str(path), 'r|*') as archive:
        for member in archive:
            name = member.name.lstrip('./')
            if not member.isfile():
                continue
            if name == 'manifest.json':
                if member.size > 8 * 1024 * 1024:
                    raise ValueError('Unreasonably large Docker save manifest')
                manifests.append(json.load(archive.extractfile(member)))
            elif member.size <= 16 * 1024 * 1024 and (name.endswith('.json') or name.startswith('blobs/sha256/')):
                raw = archive.extractfile(member).read()
                try:
                    config = json.loads(raw)
                except (UnicodeDecodeError, ValueError):
                    continue
                if isinstance(config, dict) and 'architecture' in config and 'os' in config:
                    configs[name] = {'Id': 'sha256:' + hashlib.sha256(raw).hexdigest(),
                                     'Architecture': config['architecture'], 'Os': config['os']}
    if len(manifests) != 1 or not isinstance(manifests[0], list):
        raise ValueError('Expected one Docker save manifest.json')
    result = []
    for item in manifests[0]:
        key = item['Config'].lstrip('./')
        if key not in configs:
            raise ValueError('Docker save config is missing: ' + key)
        result.append(dict(configs[key], RepoTags=item.get('RepoTags') or []))
    return result


def all_image_refs(value):
    if isinstance(value, dict):
        for key, child in value.items():
            if key == 'image' and isinstance(child, str):
                yield child
            else:
                yield from all_image_refs(child)
    elif isinstance(value, list):
        for child in value:
            yield from all_image_refs(child)


def validate_rendered(lock, compose_json, helm_yaml):
    """Validate actual service/resource semantics, not only the set of image names."""
    roles = validate_lock(lock)
    seen = {channel: set() for channel in CHANNELS}

    def check(channel, role, ref):
        expected = roles[role]['reference']
        if ref != expected:
            raise ValueError('%s role %s image mismatch: expected %s, got %s' %
                             (channel, role, expected, ref))
        seen[channel].add(role)

    compose = json.loads(compose_json) if isinstance(compose_json, str) else compose_json
    services = compose.get('services', {})
    if set(services) != set(SERVICE_ROLES):
        raise ValueError('Compose service graph differs from the release contract')
    for service, role in SERVICE_ROLES.items():
        if services[service].get('build'):
            raise ValueError('Release Compose cannot require a local image build')
        check('compose', role, services[service].get('image'))
    documents = list(yaml.safe_load_all(helm_yaml)) if isinstance(helm_yaml, str) else helm_yaml
    observed = set()
    for document in documents:
        if not isinstance(document, dict):
            continue
        observed.update(all_image_refs(document))
        kind = document.get('kind')
        spec = document.get('spec', {})
        if kind == 'OBCluster':
            check('helm', 'observer', spec.get('observer', {}).get('image'))
            if spec.get('monitor') or any(zone.get('monitor') for zone in spec.get('topology', [])):
                raise ValueError('Unvalidated optional obagent monitor is enabled')
            for zone in spec.get('topology', []):
                if (zone.get('observer') or {}).get('image'):
                    check('helm', 'observer', zone['observer']['image'])
        elif kind == 'OBLogServiceCluster':
            check('helm', 'logservice', spec.get('logService', {}).get('image'))
        elif kind in ('Deployment', 'StatefulSet', 'Job', 'CronJob', 'DaemonSet', 'Pod'):
            if kind == 'CronJob':
                spec = spec.get('jobTemplate', {}).get('spec', {})
            if kind != 'Pod':
                spec = spec.get('template', {}).get('spec', {})
            for container in spec.get('containers', []) + spec.get('initContainers', []):
                name = container.get('name')
                if name not in CONTAINER_ROLES:
                    raise ValueError('Unclassified runtime container: ' + str(name))
                check('helm', CONTAINER_ROLES[name], container.get('image'))
        elif kind == 'ConfigMap':
            embedded = (document.get('data') or {}).get('config.yaml')
            if isinstance(embedded, str):
                config = yaml.safe_load(embedded) or {}
                inspection = config.get('inspection', {})
                if inspection:
                    for key, role in (('obdiag', 'obdiag'), ('oceanbase-helper', 'oceanbase-helper')):
                        check('helm', role, inspection.get(key, {}).get('image'))
                # The optional SQL analyzer is an inactive configuration default,
                # not a delivered/validated feature. No Pod can use this exception.
                for ref in all_image_refs(config):
                    if ref in set(all_image_refs(inspection)):
                        observed.add(ref)
                    elif ref != lock.get('excluded_optional', {}).get('sql-analyzer', {}).get('reference'):
                        raise ValueError('Unclassified embedded runtime image: ' + ref)
    locked = {roles[role]['reference'] for role in HELM_ROLES}
    if observed - locked:
        raise ValueError('Untracked rendered runtime images: ' + ', '.join(sorted(observed - locked)))
    for channel in CHANNELS:
        if seen[channel] != CHANNELS[channel]:
            raise ValueError('Rendered %s is missing required image roles: %s' %
                             (channel, ', '.join(sorted(CHANNELS[channel] - seen[channel]))))
    return {channel: {role: roles[role]['reference'] for role in sorted(seen[channel])}
            for channel in CHANNELS}


def image_values(lock):
    return image_values_from_refs({role: item['reference'] for role, item in validate_lock(lock).items()})


def image_values_from_refs(r):
    """Apply one explicit role-to-values mapping in every Helm delivery builder."""
    return {'native': {'enabled': True, 'readiness': {'image': r['operator']}},
            'observer': {'image': r['observer']}, 'logService': {'image': r['logservice']},
            'minio': {'image': r['minio']}, 'initialization': {'image': r['mc']},
            'dashboard': {'image': r['dashboard'], 'utilityImage': r['utility'],
                          'prometheus': {'image': r['prometheus']},
                          'alertmanager': {'image': r['alertmanager']},
                          'inspection': {'obdiag': {'image': r['obdiag']},
                                         'oceanbaseHelper': {'image': r['oceanbase-helper']}}}}


def merge(target, overlay):
    for key, value in overlay.items():
        if isinstance(value, dict):
            merge(target.setdefault(key, {}), value)
        else:
            target[key] = value
    return target


def compose_env(lock):
    roles = validate_lock(lock)
    publication = ('Registry publication and pull are verified; see image-lock.json.'
                   if all(i.get('publication') == 'verified-registry-pull' for i in roles.values())
                   else 'Registry publication is not verified.')
    lines = ['# Generated from release-images.lock.json. ' + publication,
             '# Set IMAGE_REGISTRY only to a registry/namespace containing these exact builds.',
             'IMAGE_REGISTRY=', '',
             '# SQL, obshell Dashboard, Prometheus and Alertmanager are published by default.',
             '# Empty auto-detects the default-route source IP; set this for multi-NIC/NAT hosts.',
             'HOST_ACCESS_IP=',
             'HOST_BIND_ADDRESS=0.0.0.0',
             'OBSERVER_SQL_BIND_ADDRESS=0.0.0.0',
             'OBSERVER_SQL_HOST_PORT=32881',
             'OBSHELL_BIND_ADDRESS=0.0.0.0',
             'OBSHELL_HOST_PORT=32886',
             'AUTO_REGISTER_MONITORING=true',
             'PROMETHEUS_INTERNAL_URL=http://prometheus:9090/',
             'ALERTMANAGER_INTERNAL_URL=http://alertmanager:9093',
             'MONITORING_REGISTER_TIMEOUT_SECONDS=600']
    for key, role in ENV_ROLES.items():
        ref = roles[role]['reference']
        prefix, image = ref.rsplit('/', 1) if '/' in ref else ('docker.io/library', ref)
        lines.append('%s=${IMAGE_REGISTRY:-%s}/%s' % (key, prefix, image))
    return '\n'.join(lines) + '\n'


def write_json(path, value):
    Path(path).write_text(json.dumps(value, indent=2) + '\n')


def write_yaml(path, value):
    Path(path).write_text(yaml.safe_dump(value, sort_keys=False))


def release_timestamp(value):
    """A stable, representable date without tar's epoch/future timestamp warnings."""
    value = str(value)
    if not re.fullmatch(r'[0-9]+', value):
        raise ValueError('Release timestamp must be an integer Unix epoch')
    timestamp = int(value)
    if not 946684800 <= timestamp <= int(time.time()) + 60:
        raise ValueError('Release timestamp must be between 2000-01-01 and the current time')
    return timestamp


def tar_tree(root, target, timestamp=946684800):
    timestamp = release_timestamp(timestamp)
    with target.open('xb') as raw:
        with gzip.GzipFile(filename='', mode='wb', fileobj=raw, mtime=0, compresslevel=6) as compressed:
            with tarfile.open(fileobj=compressed, mode='w|') as archive:
                for path in [root] + sorted(root.rglob('*')):
                    if path.is_symlink():
                        raise ValueError('Symlinks are not allowed in release inputs: ' + str(path))
                    info = archive.gettarinfo(str(path), root.name + ('/' + path.relative_to(root).as_posix() if path != root else ''))
                    info.uid = info.gid = 0
                    info.mtime = timestamp
                    info.uname = info.gname = ''
                    if path.is_file():
                        with path.open('rb') as stream:
                            archive.addfile(info, stream)
                    else:
                        archive.addfile(info)


def copy_chart(source, target):
    def ignore(directory, names):
        return [name for name in names if name in ('install.sh', 'testdata', 'tests', '.git', '__pycache__')
                or name.endswith(('_test.go', '.pyc', '.tgz'))]
    shutil.copytree(source, target, ignore=ignore)


def clean_compose_env():
    # Build-time host overrides must not silently change the release contents.
    return {key: value for key, value in os.environ.items()
            if key not in ENV_ROLES and key != 'IMAGE_REGISTRY' and not key.startswith('COMPOSE_')}


def require_legacy_compose_source(source):
    """The shared-image bundle builder cannot select Compose wrapper images."""
    compose = yaml.safe_load((source / 'deploy/compose/lakehouse/compose.yaml').read_text())

    def embedded_runtime(value):
        if isinstance(value, str):
            return '/opt/lakehouse/' in value
        if isinstance(value, dict):
            return any(embedded_runtime(child) for child in value.values())
        if isinstance(value, list):
            return any(embedded_runtime(child) for child in value)
        return False

    def script_bind(service):
        for mount in service.get('volumes') or []:
            if isinstance(mount, dict):
                if mount.get('type') == 'bind' and str(mount.get('target', '')).rstrip('/') == '/opt/lakehouse':
                    return True
            elif isinstance(mount, str):
                parts = mount.split(':')
                if (len(parts) >= 2 and parts[0].startswith(('.', '/', '~'))
                        and parts[1].rstrip('/') == '/opt/lakehouse'):
                    return True
        return False

    for service in (compose.get('services') or {}).values():
        if (any(embedded_runtime(service.get(field)) for field in ('entrypoint', 'command', 'healthcheck'))
                and not script_bind(service)):
            raise ValueError(
                'Legacy combined builder cannot package embedded-runtime Compose with shared base images; '
                'use build-compose-release.py and build-helm-release.py for channel-specific artifacts')


def build(args):
    source = Path(args.source).resolve()
    require_legacy_compose_source(source)
    lock = load_lock(args.lock)
    roles = validate_lock(lock)
    publication_evidence = None
    if getattr(args, 'publication_report', None):
        publication = json.loads(Path(args.publication_report).read_text())
        for role, item in roles.items():
            proof = publication.get(role, {})
            if (not proof.get('registry_pull_verified') or
                    any(proof.get(key) != item.get(key) for key in
                        ('reference', 'image_id', 'platform', 'manifest_digest')) or
                    not IMAGE_ID.fullmatch(proof.get('manifest_digest', ''))):
                raise ValueError('Registry publication evidence mismatch: ' + role)
        publication_evidence = {'source': 'supplied-push-and-pull-report',
                                'sha256': sha256(args.publication_report),
                                'images_checked': len(roles)}
    if not re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9._-]{0,79}', args.version):
        raise ValueError('Invalid delivery version')
    output = Path(args.output).resolve()
    if output.exists():
        raise ValueError('Output already exists; choose a new directory')
    if getattr(args, 'configuration_only', False):
        evidence = {'kind': 'configuration-only', 'images_checked': 0,
                    'live_docker_checked': False, 'image_layers_verified': False,
                    'registry_pull_verified': False, 'archive_config_bytes_verified': False}
    elif args.docker_metadata:
        evidence = verify_metadata(lock, json.loads(Path(args.docker_metadata).read_text()))
        evidence.update(source='supplied-metadata-snapshot', sha256=sha256(args.docker_metadata),
                        live_docker_checked=False)
    elif args.image_archive:
        metadata = [item for path in args.image_archive for item in archive_metadata(path)]
        evidence = verify_metadata(lock, metadata)
        evidence.update(kind='docker-save-config-bytes', archive_config_bytes_verified=True,
                        live_docker_checked=False,
                        archives=[{'file': Path(path).name, 'sha256': sha256(path)} for path in args.image_archive])
    else:
        metadata = json.loads(run([args.docker, 'image', 'inspect'] + [i['reference'] for i in roles.values()]))
        evidence = verify_metadata(lock, metadata)
        evidence.update(source='live-docker-inspect', live_docker_checked=True)
    source_commit = run(['git', 'rev-parse', 'HEAD'], cwd=source).strip()
    source_dirty = bool(run(['git', 'status', '--porcelain'], cwd=source).strip())
    timestamp = release_timestamp(os.environ.get('SOURCE_DATE_EPOCH') or
                                  run(['git', 'show', '-s', '--format=%ct', 'HEAD'], cwd=source).strip())
    output.mkdir(parents=True)
    compose = output / ('oceanbase-lakehouse-compose-' + args.version)
    helm = output / ('oceanbase-lakehouse-native-helm-' + args.version)
    compose.mkdir(); helm.mkdir()
    for name in ('compose.yaml', 'README.md'):
        shutil.copyfile(source / 'deploy/compose/lakehouse' / name, compose / name)
    shutil.copytree(source / 'deploy/compose/lakehouse/scripts', compose / 'scripts')
    shutil.copytree(source / 'deploy/compose/lakehouse/monitoring', compose / 'monitoring')
    (compose / '.env').write_text(compose_env(lock))
    values = image_values(lock)
    write_yaml(helm / 'image-values.yaml', values)
    operator_values = {'image': roles['operator']['reference'],
                       'kubeRbacProxyImage': roles['kube-rbac-proxy']['reference']}
    write_yaml(helm / 'operator-values.yaml', operator_values)
    with tempfile.TemporaryDirectory(prefix='lakehouse-release-charts-') as temp:
        temp = Path(temp)
        for name in ('oceanbase-dashboard', 'oceanbase-minio', 'oceanbase-lakehouse', 'ob-operator'):
            copy_chart(source / 'charts' / name, temp / name)
        parent = temp / 'oceanbase-lakehouse'
        write_yaml(parent / 'values.yaml', merge(yaml.safe_load((parent / 'values.yaml').read_text()), values))
        op = temp / 'ob-operator'
        write_yaml(op / 'values.yaml', merge(yaml.safe_load((op / 'values.yaml').read_text()),
                                            operator_values))
        (parent / 'charts').mkdir(exist_ok=True)
        for name in ('oceanbase-dashboard', 'oceanbase-minio'):
            child = temp / name
            child_version = yaml.safe_load((child / 'Chart.yaml').read_text())['version']
            tar_tree(child, parent / 'charts' / ('%s-%s.tgz' % (name, child_version)), timestamp)
        for chart in (parent, op):
            run([args.helm, 'lint', '--strict', chart])
            run([args.helm, 'package', chart, '--destination', helm])
    for name in ('values-single-node.yaml', 'values-three-nodes.yaml', 'USAGE.txt', 'SHARED-INFRA.md'):
        shutil.copyfile(source / 'charts/oceanbase-lakehouse' / name, helm / name)
    lakehouse_chart, = helm.glob('oceanbase-lakehouse-*.tgz')
    operator_chart, = helm.glob('ob-operator-*.tgz')
    (helm / 'QUICKSTART.md').write_text(
        '# Install Lakehouse with Dashboard\n\n'
        'Use a prepared Kubernetes cluster with the matching OceanBase Operator, '
        'cert-manager, a working `local-path` StorageClass, and the release images '
        'available on its workers. See `USAGE.txt` for administrator preparation.\n\n'
        'From this extracted directory, run one command:\n\n```bash\n'
        'helm upgrade --install ob-demo ./' + lakehouse_chart.name +
        ' --namespace ob-demo --create-namespace --wait --wait-for-jobs --timeout 35m\n'
        '```\n\n'
        'Dashboard and its readiness check are enabled by default. No values file '
        'or separate Dashboard installation is required. The default experience '
        'uses one Observer and three LogService processes.\n\n'
        'These are first-install defaults. Keep the existing sizing and custom '
        'values when upgrading an installation.\n\n'
        'For another StorageClass, append `--set storageClass=YOUR_STORAGE_CLASS --set minio.storageClass=YOUR_STORAGE_CLASS`. '
        'Dashboard and SQL use automatically allocated NodePorts. '
        'Helm prints the node IP and login commands. For NAT or multiple NICs, '
        'append `--set access.host=YOUR_REACHABLE_IP`.\n\n'
        'Optional image acceleration: [registry-mirrors/README.md](registry-mirrors/README.md). '
        'It preserves image names and is configured on workers, separately from Helm. '
        'The examples are not applied automatically and do not publish missing images.\n')
    compose_json = run([args.docker, 'compose', '--env-file', compose / '.env', '-f', compose / 'compose.yaml',
                        'config', '--format', 'json'], cwd=compose, env=clean_compose_env())
    op_render = run([args.helm, 'template', 'ob-operator', operator_chart, '-n', 'oceanbase-system'])
    rendered = {}
    for profile in (None, 'values-single-node.yaml', 'values-three-nodes.yaml'):
        command = [args.helm, 'template', 'release', lakehouse_chart, '-n', 'release-check']
        if profile:
            command += ['-f', helm / profile]
        rendered[profile or 'defaults'] = validate_rendered(lock, compose_json, run(command) + '\n---\n' + op_render)
    report = {'source_commit': source_commit, 'source_worktree_dirty': source_dirty,
              'archive_timestamp': timestamp,
              'release_lock_sha256': sha256(args.lock), 'identity_evidence': evidence,
              'rendered_roles': rendered, 'live_installation_tested': False,
              'registry_publication_verified': publication_evidence is not None,
              'registry_publication_evidence': publication_evidence,
              'image_payload_included': False}
    for channel, directory in (('compose', compose), ('helm', helm)):
        (directory / 'tools').mkdir()
        shutil.copyfile(HERE / ('package-offline-' + channel + '.py'), directory / 'tools/package-offline.py')
        shutil.copyfile(HERE / 'offline_common.py', directory / 'tools/offline_common.py')
        if channel == 'helm':
            shutil.copyfile(HERE / 'migrate-minio-values.py', directory / 'tools/migrate-minio-values.py')
        (directory / 'tools/requirements.txt').write_text('PyYAML==6.0.3\n')
        shutil.copyfile(HERE / 'OFFLINE-LEGACY.md', directory / 'OFFLINE.md')
        shutil.copyfile(HERE / 'THIRD-PARTY-NOTICES.md', directory / 'THIRD-PARTY-NOTICES.md')
        if channel == 'compose':
            shutil.copyfile(directory / '.env', directory / '.env.example')
        else:
            shutil.copyfile(directory / 'QUICKSTART.md', directory / 'README.md')
        with (directory / 'README.md').open('a') as readme:
            readme.write('\n## Offline preparation and third-party components\n\n'
                         'See [OFFLINE.md](OFFLINE.md) for image listing, checking and offline packaging. '
                         'See [THIRD-PARTY-NOTICES.md](THIRD-PARTY-NOTICES.md) for MinIO/mc notices.\n'
                         'This configuration bundle contains no image layers. ' +
                         ('All release images have push/pull evidence for the configured registry.\n'
                          if publication_evidence else
                          'Internal image publication/public pull access is not verified.\n'))
        shutil.copytree(source / 'deploy/registry-mirrors', directory / 'registry-mirrors',
                        ignore=shutil.ignore_patterns('tests', '__pycache__', '*.pyc'))
        selected = copy.deepcopy(lock)
        selected['images'] = [i for i in selected['images'] if channel in i['channels']]
        selected.update(delivery_channel=channel, delivery_version=args.version, releaseVersion=args.version)
        write_json(directory / 'image-lock.json', selected)
        shutil.copyfile(args.lock, directory / 'release-images.lock.json')
        shutil.copyfile(source / 'LICENSE', directory / 'LICENSE')
        write_json(directory / 'build-validation.json', report)
        (directory / 'IMAGE-DELIVERY.txt').write_text(
            'Images are not included in this bundle. Canonical image references and config IDs are in image-lock.json.\n'
            'source_reference records build-cache provenance only. It is not a published registry promise.\n'
            'Prepare the exact canonical tags in your runtime, or mirror identical images to your registry.\n' +
            ('Registry publication and pull verified against the supplied report. Target network access is required.\n'
             if publication_evidence else
             'Registry publication and public pull access have NOT been verified by this package build.\n') +
            'Optional SQL analyzer is not included or validated; cert-manager is an external prerequisite.\n'
            'Compose includes the open-source MinIO/mc references, OBAgent, Prometheus and Alertmanager.\n')
        checksums = [sha256(p) + '  ' + p.relative_to(directory).as_posix() for p in sorted(directory.rglob('*')) if p.is_file()]
        (directory / 'SHA256SUMS').write_text('\n'.join(checksums) + '\n')
        tar_tree(directory, output / (directory.name + '.tar.gz'), timestamp)
    report['artifacts'] = [{'file': path.name, 'size_bytes': path.stat().st_size, 'sha256': sha256(path)}
                           for path in sorted(output.glob('*.tar.gz'))]
    write_json(output / 'build-result.json', report)
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--source', type=Path, default=HERE.parent.parent)
    parser.add_argument('--lock', type=Path, default=HERE / 'release-images.lock.json')
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--version', required=True)
    parser.add_argument('--publication-report', type=Path,
                        help='Actual per-role registry push/pull evidence, matching the lock')
    evidence = parser.add_mutually_exclusive_group()
    evidence.add_argument('--configuration-only', action='store_true',
                          help='Build configuration preview only; explicitly skip image identity evidence (not a verified image release)')
    evidence.add_argument('--docker-metadata', type=Path, help='Saved docker image inspect JSON; only metadata is verified')
    evidence.add_argument('--image-archive', type=Path, action='append', help='Docker save tar/tar.gz with canonical tags; config bytes are verified')
    parser.add_argument('--helm', default='helm')
    parser.add_argument('--docker', default='docker')
    args = parser.parse_args()
    try:
        print(json.dumps(build(args), indent=2))
    except (ValueError, OSError, KeyError, subprocess.SubprocessError) as error:
        parser.exit(1, 'Native package build failed: %s\n' % error)


if __name__ == '__main__':
    main()

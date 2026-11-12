#!/usr/bin/env python3
"""Build a standard, source-derived development Helm chart from a reviewed lock.

No baseline chart is read, no image is built/pulled, and no cluster is contacted.
The raw source, versioned publishing profile and image lock are snapshotted so
all differences between raw-source and published defaults are explicit.
"""
import argparse
import copy
import difflib
import gzip
import hashlib
import importlib.util
import json
from pathlib import Path
import platform
import shutil
import subprocess
import tarfile
import tempfile

import yaml

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location('native_release', HERE / 'build-native-packages.py')
native = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(native)
CHART_NAME = 'oceanbase-lakehouse'
OPERATOR_BOOTSTRAP_ROLES = {'operator', 'kube-rbac-proxy'}
CERT_MANAGER_ROLES = {'cert-manager-controller', 'cert-manager-cainjector',
                      'cert-manager-webhook', 'cert-manager-acmesolver'}
REQUIRED_ROLES = native.HELM_ROLES | CERT_MANAGER_ROLES
PROFILE = Path('scripts/lakehouse/release-profiles/single-command.yaml')
INPUT_PATHS = [Path('charts') / name for name in
               ('oceanbase-lakehouse', 'oceanbase-dashboard', 'oceanbase-minio')]
INPUT_PATHS += [Path('scripts/lakehouse/prerequisites'),
                Path('scripts/lakehouse/images/minio-rejoin'), PROFILE]
INPUT_PATHS += [Path('scripts/lakehouse') / name for name in
                ('build-helm-release.py', 'build-native-packages.py', 'build-one-command-package.py',
                 'install-helm-and-show-access.sh',
                 'migrate-minio-values.py', 'progress/progress.py', 'progress/progress.sh')]


def validate_lock(lock):
    if lock.get('platform') != 'linux/amd64':
        raise ValueError('This development release is validated only for linux/amd64')
    roles = {}
    refs = set()
    for item in lock.get('images', []):
        role = item.get('role')
        ref = item.get('reference')
        if role in roles or not role:
            raise ValueError('Missing or duplicate image role: ' + str(role))
        if not isinstance(ref, str) or not ref or ref in refs:
            raise ValueError('Missing or duplicate image reference: ' + str(ref))
        if not native.IMAGE_ID.fullmatch(item.get('image_id', '')):
            raise ValueError('Missing immutable config ID for ' + role)
        if not native.IMAGE_ID.fullmatch(item.get('manifest_digest', '')):
            raise ValueError('Missing registry manifest digest for ' + role)
        if item.get('platform') != lock['platform']:
            raise ValueError('Image platform mismatch: ' + role)
        roles[role] = item
        refs.add(ref)
    missing = REQUIRED_ROLES - roles.keys()
    if missing:
        raise ValueError('Missing Helm image roles: ' + ', '.join(sorted(missing)))
    return roles


def source_snapshot(source, destination):
    """Copy only declared publishing inputs; build from this frozen copy."""
    destination.mkdir()
    for relative in INPUT_PATHS:
        original = source / relative
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        if original.is_dir():
            native.copy_chart(original, target)
        else:
            shutil.copy2(original, target)
    files = {p.relative_to(destination).as_posix():
             {'sha256': native.sha256(p), 'mode': p.stat().st_mode & 0o777}
             for p in sorted(destination.rglob('*')) if p.is_file()}
    tree_digest = hashlib.sha256(json.dumps(files, sort_keys=True).encode()).hexdigest()
    return files, tree_digest


def source_provenance(source, used_files):
    """The tree snapshot is authoritative even when building exported sources."""
    def git(*args):
        p = subprocess.run(['git', *args], cwd=source, text=True, capture_output=True)
        return p.stdout if not p.returncode else None
    commit = git('rev-parse', 'HEAD')
    if commit is None:
        return {'source_commit': None, 'source_worktree_dirty': None,
                'git_metadata_available': False}, ''
    scopes = [p.as_posix() for p in INPUT_PATHS]
    changed = (git('diff', '--name-only', 'HEAD', '--', *scopes) or '').splitlines()
    # Ignore generated archives and test fixtures which the input snapshot does
    # not consume, but retain deletions of source runtime files.
    changed = [name for name in changed if name in used_files or (
        not (source / name).exists() and not name.endswith(('_test.go', '.tgz', '.pyc'))
        and not {'tests', 'testdata', '__pycache__'}.intersection(Path(name).parts))]
    patch = (git('diff', 'HEAD', '--', *changed) or '') if changed else ''
    for name in (git('ls-files', '--others', '--exclude-standard', '--', *scopes) or '').splitlines():
        path = source / name
        if name in used_files and path.is_file():
            patch += 'diff --git a/%s b/%s\nnew file mode %06o\n' % (name, name, path.stat().st_mode & 0o100777)
            patch += ''.join(difflib.unified_diff([], path.read_text().splitlines(True),
                                                fromfile='/dev/null', tofile='b/' + name))
    return {'source_commit': commit.strip(),
            'source_worktree_dirty': bool(git('status', '--porcelain')),
            'git_metadata_available': True}, patch


def rewrite_payload_images(path, wanted):
    """Map declared workload container names, retaining all other source bytes."""
    text = path.read_text()
    seen = set()
    for document in yaml.safe_load_all(text):
        if not isinstance(document, dict) or document.get('kind') != 'Deployment':
            continue
        deployment = document['metadata']['name']
        containers = document['spec']['template']['spec']['containers']
        for container in containers:
            key = (deployment, container['name'])
            if key not in wanted:
                raise ValueError('Unmapped prerequisite workload container: ' + repr(key))
            old = container['image']
            new = wanted[key]
            # Existing manifests use plain YAML image scalars. Parse first so
            # absent/renamed containers cannot silently keep an old image.
            original = 'image: ' + old
            if text.count(original) != 1:
                raise ValueError('Expected one plain image scalar in ' + str(path))
            text = text.replace(original, 'image: ' + new, 1)
            seen.add(key)
    if seen != set(wanted):
        raise ValueError('Missing prerequisite containers: ' + repr(set(wanted) - seen))
    return text


def write_foundation(chart, cert_manager, operator, operator_crds, roles):
    """Separate CRD registration from the ordered Helm hooks that use those APIs."""
    crds = chart / 'crds'
    crds.mkdir()
    for path in operator_crds:
        shutil.copy2(path, crds / path.name)
    cert_objects = [obj for obj in yaml.safe_load_all(cert_manager.read_bytes()) if obj]
    images = {}
    for obj in cert_objects:
        if obj['kind'] == 'CustomResourceDefinition':
            native.write_yaml(crds / (obj['metadata']['name'] + '.yaml'), obj)
        if obj['kind'] == 'Deployment':
            for container in obj['spec']['template']['spec']['containers']:
                images[container['name']] = container['image']
                for arg in container.get('args', []):
                    if arg.startswith('--acme-http01-solver-image='):
                        images['cert-manager-acmesolver'] = arg.split('=', 1)[1]
    if images != {role: roles[role]['reference'] for role in CERT_MANAGER_ROLES}:
        raise ValueError('Bundled cert-manager images must match the Helm image lock')
    native.write_json(chart / 'files/prerequisites/foundation.json', {
        'certManager': [obj for obj in cert_objects if obj['kind'] != 'CustomResourceDefinition'],
        'operator': [
            {'apiVersion': 'v1', 'kind': 'Namespace', 'metadata': {'name': 'oceanbase-system'}},
            *[obj for obj in yaml.safe_load_all(operator) if obj],
        ],
    })


def assemble_chart(snapshot, lock, destination):
    """Construct the complete chart from current source, never a baseline tgz."""
    roles = validate_lock(lock)
    parent_source = snapshot / 'charts' / CHART_NAME
    native.copy_chart(parent_source, destination)
    meta = yaml.safe_load((destination / 'Chart.yaml').read_text())
    if meta['name'] != CHART_NAME:
        raise ValueError('Unexpected source Chart name')
    charts = destination / 'charts'
    shutil.rmtree(charts, ignore_errors=True)
    charts.mkdir()
    for dependency in meta.get('dependencies', []):
        repository = dependency.get('repository', '')
        if not repository.startswith('file://'):
            raise ValueError('Build requires vendored source for dependency ' + dependency['name'])
        original = (parent_source / repository[len('file://'):]).resolve()
        original.relative_to((snapshot / 'charts').resolve())
        child = destination.parent / dependency['name']
        native.copy_chart(original, child)
        child_meta = yaml.safe_load((child / 'Chart.yaml').read_text())
        if (child_meta['name'], child_meta['version']) != (dependency['name'], dependency['version']):
            raise ValueError('Source dependency name/version differs from Chart.yaml')
        native.tar_tree(child, charts / ('%s-%s.tgz' % (child_meta['name'], child_meta['version'])))
    prerequisites = snapshot / 'scripts/lakehouse/prerequisites'
    payload = destination / 'files/prerequisites'
    payload.mkdir(parents=True)
    operator = rewrite_payload_images(prerequisites / 'manifests/operator.yaml', {
        ('oceanbase-controller-manager', 'manager'): roles['operator']['reference'],
        ('oceanbase-controller-manager', 'kube-rbac-proxy'): roles['kube-rbac-proxy']['reference'],
    })
    profile = yaml.safe_load((snapshot / PROFILE).read_text())
    if profile.get('schemaVersion') != 1 or profile.get('name') != 'single-command':
        raise ValueError('Unknown single-command profile')
    image_values = native.image_values_from_refs({r: i['reference'] for r, i in roles.items()})
    overlay = native.merge(copy.deepcopy(profile['values']), image_values)
    cert_manager = prerequisites / 'cert-manager-locked.yaml'
    if not cert_manager.is_file():
        raise ValueError('Provide the pinned --cert-manager-manifest for the one-command Chart')
    operator_crds = sorted((prerequisites / 'crds').glob('*.oceanbase.com.yaml'))
    write_foundation(destination, cert_manager, operator, operator_crds, roles)
    shutil.copy2(prerequisites / 'wait-ready.py', payload / 'wait-ready.py')
    overlay['prerequisites']['image'] = roles['oceanbase-helper']['reference']
    values = native.merge(yaml.safe_load((destination / 'values.yaml').read_text()), overlay)
    native.write_yaml(destination / 'values.yaml', values)
    return meta, overlay


def operator_bootstrap_lock(lock, roles):
    return {'schemaVersion': 1, 'platform': lock['platform'],
            'delivery_channel': 'operator-bootstrap',
            'images': [roles[role] for role in sorted(OPERATOR_BOOTSTRAP_ROLES)]}


def package_chart(chart, output):
    """Use Helm ignore rules/metadata, then normalize archive headers."""
    meta = yaml.safe_load((chart / 'Chart.yaml').read_text())
    name = '%s-%s.tgz' % (meta['name'], meta['version'])
    target = output / name
    with tempfile.TemporaryDirectory() as tmp:
        native.run(['helm', 'package', chart, '--destination', tmp])
        with tarfile.open(Path(tmp) / name) as original, target.open('xb') as raw:
            with gzip.GzipFile(filename='', fileobj=raw, mode='wb', mtime=0) as compressed:
                with tarfile.open(fileobj=compressed, mode='w|') as result:
                    for member in sorted(original.getmembers(), key=lambda m: m.name):
                        if member.issym() or member.islnk() or not (member.isfile() or member.isdir()):
                            raise ValueError('Unexpected chart archive member')
                        member.uid = member.gid = 0
                        member.uname = member.gname = ''
                        member.mtime = 946684800
                        result.addfile(member, original.extractfile(member) if member.isfile() else None)
    return target


def build_release(source, lock, output, release_version=None, allow_existing_output=False,
                  *, cert_manager=None):
    source, output = Path(source).resolve(), Path(output).resolve()
    roles = validate_lock(lock)
    for name in ('build-helm-release.py', 'build-native-packages.py', 'build-one-command-package.py'):
        if (source / 'scripts/lakehouse' / name).read_bytes() != (HERE / name).read_bytes():
            raise ValueError('Invoke the publishing builder from the selected source tree: ' + name)
    if output.exists() and not allow_existing_output:
        raise ValueError('Output exists; choose a new directory')
    output.mkdir(parents=True, exist_ok=allow_existing_output)
    provenance = output / 'provenance'
    if provenance.exists():
        shutil.rmtree(provenance)
    provenance.mkdir()
    snapshot = provenance / 'source-inputs'
    files, tree_digest = source_snapshot(source, snapshot)
    if cert_manager is not None:
        pinned = snapshot / 'scripts/lakehouse/prerequisites/cert-manager-locked.yaml'
        shutil.copyfile(cert_manager, pinned)
        files[pinned.relative_to(snapshot).as_posix()] = {'sha256': native.sha256(pinned), 'mode': pinned.stat().st_mode & 0o777}
        tree_digest = hashlib.sha256(json.dumps(files, sort_keys=True).encode()).hexdigest()
    metadata, patch = source_provenance(source, files)
    (provenance / 'source.patch').write_text(patch)
    native.write_json(provenance / 'source-tree.json', files)
    native.write_json(provenance / 'input-image-lock.json', lock)
    with tempfile.TemporaryDirectory(prefix='lakehouse-source-chart-') as tmp:
        chart = Path(tmp) / CHART_NAME
        meta, overlay = assemble_chart(snapshot, lock, chart)
        native.run(['helm', 'lint', '--strict', chart])
        rendered = native.run(['helm', 'template', 'ob-demo', chart, '-n', 'ob-demo', '--include-crds'])
        (output / 'rendered-validation.yaml').write_text(rendered)
        profiles = ('values-three-nodes.yaml', 'values-minio-ha.yaml')
        profile_flags = [arg for name in profiles for arg in ('-f', chart / name)]
        native.run(['helm', 'template', 'ob-demo', chart, '-n', 'ob-demo', *profile_flags])
        target = package_chart(chart, output)
        for name in profiles:
            shutil.copyfile(chart / name, output / name)
    native.write_yaml(output / 'release-defaults.yaml', overlay)
    release_version = release_version or meta['version']
    effective_lock = copy.deepcopy(lock)
    effective_lock.update(delivery_channel='helm', delivery_version=release_version, releaseVersion=release_version)
    effective_lock['images'] = [roles[role] for role in sorted(REQUIRED_ROLES)]
    effective_lock['prerequisites'] = [
        'Kubernetes, Helm and a working StorageClass',
        'Cluster-level install permissions for the bundled CRDs and shared components',
    ]
    native.write_json(output / 'image-lock.json', effective_lock)
    native.write_json(output / 'operator-bootstrap-image-lock.json', operator_bootstrap_lock(lock, roles))
    native.tar_tree(snapshot, provenance / 'source-inputs.tar.gz')
    inputs = {**metadata, 'chart_name': meta['name'], 'chart_version': meta['version'],
              'app_version': meta.get('appVersion'), 'dependencies': meta.get('dependencies', []),
              'profile': PROFILE.as_posix(), 'source_tree_sha256': tree_digest,
              'source_snapshot_sha256': native.sha256(provenance / 'source-inputs.tar.gz'),
              'source_patch_sha256': native.sha256(provenance / 'source.patch'),
              'input_lock_sha256': native.sha256(provenance / 'input-image-lock.json'),
              'release_defaults_sha256': native.sha256(output / 'release-defaults.yaml'),
              'chart_sha256': native.sha256(target),
              'image_roles': {role: {key: roles[role][key] for key in
                                    ('reference', 'platform', 'image_id', 'manifest_digest')}
                              for role in sorted(REQUIRED_ROLES)},
              'raw_source_values_modified': False, 'baseline_chart_used': False,
              'tool_versions': {'helm': native.run(['helm', 'version', '--template', '{{.Version}}']).strip(),
                                'python': platform.python_version(), 'pyyaml': yaml.__version__},
              'archive_timestamp': 946684800}
    native.write_json(output / 'release-inputs.json', inputs)
    native.write_json(output / 'build-validation.json', {
        'version': release_version, 'chart_version': meta['version'], 'helm_lint': True,
        'helm_template': True, 'live_business_installation_tested': False,
        'registry_images': len(REQUIRED_ROLES), 'registry_pull_reverified_by_this_build': False,
        'source_tree_sha256': tree_digest, 'chart_sha256': inputs['chart_sha256']})
    return {'chart': str(target), 'chart_version': meta['version'],
            'chart_sha256': inputs['chart_sha256'], 'source_tree_sha256': tree_digest,
            'release_inputs': str(output / 'release-inputs.json')}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--source', type=Path, default=HERE.parent.parent)
    parser.add_argument('--lock', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--release-version', help='Delivery label only; Chart version always comes from source Chart.yaml')
    parser.add_argument('--cert-manager-manifest', type=Path, help='Pinned cert-manager YAML to bundle in the chart')
    args = parser.parse_args()
    result = build_release(args.source, json.loads(args.lock.read_text()), args.output, args.release_version,
                           cert_manager=args.cert_manager_manifest)
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()

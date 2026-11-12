#!/usr/bin/env python3
"""Prepare locked release images using the delivery-specific importer.

Python 3.8+, PyYAML, Docker Compose (Compose bundles), Helm (Helm bundles).
Tar mode requires skopeo and never imports images into local Docker.
Does not deploy services or push images. Import mode changes the selected
runtime's image cache. Offline Compose aliases retain locked config identities.
"""
import argparse
import copy
import gzip
import hashlib
import inspect
import io
import json
import os
from pathlib import Path
import re
import shutil
import shlex
import subprocess
import sys
import tarfile
import tempfile
import urllib.parse
import urllib.request

import yaml


SHA256 = re.compile(r'sha256:[0-9a-f]{64}')
SERVICE_ENDPOINT = re.compile(
    r'http://[a-z0-9-]+\.[a-z0-9-]+\.svc(?:\.cluster\.local)?(?::([0-9]{1,5}))?')


def run(argv):
    result = subprocess.run([str(x) for x in argv], stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE, text=True, timeout=300)
    if result.returncode:
        raise ValueError('Command failed: %s\n%s' %
                         (' '.join(map(str, argv)), result.stderr.strip()))
    return result.stdout


def refs_in(value, include_solver=False):
    if isinstance(value, dict):
        for key, child in value.items():
            if key == 'image' and isinstance(child, str):
                yield child
            else:
                yield from refs_in(child, include_solver)
    elif isinstance(value, list):
        for index, child in enumerate(value):
            if include_solver and isinstance(child, str) and child.startswith('--acme-http01-solver-image='):
                yield child.partition('=')[2]
            elif include_solver and child == '--acme-http01-solver-image' and index + 1 < len(value):
                solver = value[index + 1]
                if isinstance(solver, str):
                    yield solver
            else:
                yield from refs_in(child, include_solver)


def manifest_refs(documents, excluded, include_solver=False):
    refs = set()
    for doc in documents:
        if not isinstance(doc, dict):
            continue
        # Includes OBCluster/OBLogServiceCluster and topology overrides, not
        # only Pod images. These CR images are instantiated later by Operator.
        refs.update(refs_in(doc, include_solver))
        if doc.get('kind') == 'ConfigMap':
            for value in (doc.get('data') or {}).values():
                if not isinstance(value, str):
                    continue
                try:
                    embedded = list(yaml.safe_load_all(value))
                except yaml.YAMLError:
                    continue
                # Dashboard diagnostic jobs are created dynamically. Only
                # documented inactive ConfigMap defaults can be excluded.
                refs.update(set(refs_in(embedded, include_solver)) - excluded)
    return refs


def one_chart(bundle, pattern):
    matches = list(bundle.glob(pattern))
    if len(matches) != 1:
        raise ValueError('Expected exactly one ' + pattern)
    return matches[0]


def fetch_bytes(source, limit=64 * 1024 * 1024):
    """Read release metadata/artifacts; image downloads use the runtime tools."""
    parsed = urllib.parse.urlsplit(str(source))
    if parsed.scheme in ('http', 'https'):
        with urllib.request.urlopen(str(source), timeout=60) as response:
            content = response.read(limit + 1)
    elif parsed.scheme:
        raise ValueError('Release assets support HTTP(S) or local paths only')
    else:
        with Path(source).open('rb') as stream:
            content = stream.read(limit + 1)
    if len(content) > limit:
        raise ValueError('Release metadata/artifact exceeds the supported size')
    return content


def resolve_asset(manifest_source, asset_url):
    if not isinstance(asset_url, str) or not asset_url:
        raise ValueError('Release artifact URL/path is required')
    parent = urllib.parse.urlsplit(str(manifest_source))
    if parent.scheme in ('http', 'https'):
        resolved = urllib.parse.urljoin(str(manifest_source), asset_url)
        if urllib.parse.urlsplit(resolved).scheme not in ('http', 'https'):
            raise ValueError('A remote release cannot reference local files')
        return resolved
    if urllib.parse.urlsplit(asset_url).scheme:
        return asset_url
    return str((Path(manifest_source).resolve().parent / asset_url).resolve())


def checked_asset(source, descriptor, target):
    if not isinstance(descriptor, dict):
        raise ValueError('Release artifact requires URL and SHA256')
    expected = descriptor.get('sha256', '')
    if len(expected) != 64 or any(c not in '0123456789abcdef' for c in expected):
        raise ValueError('Release artifact requires a lowercase SHA256')
    content = fetch_bytes(resolve_asset(source, descriptor.get('url')))
    if hashlib.sha256(content).hexdigest() != expected:
        raise ValueError('Release artifact SHA256 mismatch: ' + target.name)
    target.write_bytes(content)
    target.chmod(0o600)


def manifest_bundle(args):
    source = str(args.release_manifest)
    raw = fetch_bytes(source, 4 * 1024 * 1024)
    expected = getattr(args, 'manifest_sha256', None)
    if expected and hashlib.sha256(raw).hexdigest() != expected:
        raise ValueError('Release manifest SHA256 mismatch')
    document = json.loads(raw)
    if not isinstance(document, dict) or document.get('schemaVersion') != 1:
        raise ValueError('Unsupported release manifest schemaVersion')
    channel = getattr(args, 'expected_channel', None)
    if channel not in ('compose', 'helm'):
        raise ValueError('Select --channel compose or helm for a release manifest')
    channels = document.get('channels')
    if not isinstance(channels, dict):
        raise ValueError('Release manifest requires channel descriptors')
    entry = channels.get(channel)
    if not isinstance(entry, dict):
        raise ValueError('Release manifest does not provide channel: ' + channel)
    version, platform = document.get('releaseVersion'), document.get('platform')
    if not isinstance(version, str) or not version or platform != 'linux/amd64':
        raise ValueError('Release manifest requires releaseVersion and supported linux/amd64 platform')
    images = entry.get('images')
    if not isinstance(images, list) or not images:
        raise ValueError('Release manifest requires locked channel images')
    seen = {}
    for item in images:
        if not isinstance(item, dict):
            raise ValueError('Release images must be objects')
        ref, identity = item.get('reference'), item.get('image_id', '')
        if not isinstance(ref, str) or not ref or not isinstance(identity, str) or not SHA256.fullmatch(identity):
            raise ValueError('Release images require reference and SHA256 config image_id')
        if '@' in ref:
            digest = ref.rsplit('@', 1)[1]
            if not SHA256.fullmatch(digest) or digest != item.get('manifest_digest', digest):
                raise ValueError('Digest reference differs from the locked registry manifest: ' + ref)
        if ref in seen and seen[ref] != identity:
            raise ValueError('Conflicting image identity for release reference')
        seen[ref] = identity
    context = tempfile.TemporaryDirectory(prefix='lakehouse-release-')
    args._release_context = context
    bundle = Path(context.name)
    assets = []
    required = [('configuration', 'compose.yaml')] if channel == 'compose' else [('chart', 'oceanbase-lakehouse-release.tgz')]
    if channel == 'helm':
        required += [(key, filename) for key, filename in [('operatorChart', 'ob-operator-release.tgz'), ('operatorValues', 'operator-values.yaml')] if key in entry]
        if 'operatorValues' in entry and 'operatorChart' not in entry:
            raise ValueError('operatorValues requires operatorChart')
        if getattr(args, 'include_prerequisites', False) and 'certManagerManifest' in entry:
            required.append(('certManagerManifest', 'cert-manager.yaml'))
    for key, filename in required:
        path = bundle / filename
        checked_asset(source, entry.get(key), path)
        assets.append(path)
    lock = {'delivery_channel': channel, 'delivery_version': version, 'platform': platform,
            'images': images, 'prerequisites': entry.get('prerequisites', []),
            'excluded_optional': entry.get('excludedOptional', entry.get('excluded_optional', {}))}
    args.bundle = bundle
    args._release_assets = assets
    return bundle, lock, {'schemaVersion': 1, 'releaseVersion': version,
                          'sha256': hashlib.sha256(raw).hexdigest()}


def helm_values(snapshots):
    """Read the planning fields in the same overlay order as Helm."""
    storage = {}
    skip_operator = False
    for index, data in enumerate(snapshots, 1):
        try:
            for document in yaml.safe_load_all(data):
                if not isinstance(document, dict):
                    continue
                if 'objectStorage' in document:
                    overlay = document['objectStorage']
                    if isinstance(overlay, dict):
                        storage.update(overlay)
                    else:
                        storage.clear()
                if 'skip-operator' in document:
                    skip_operator = document['skip-operator'] is True
        except yaml.YAMLError:
            # A YAML parser error can contain the original credential line.
            raise ValueError('Invalid YAML in Helm values file %d' % index) from None
    return storage.get('endpoint'), skip_operator


def helm_image_plan(args, bundle, release, excluded):
    release_name = getattr(args, 'release', None)
    namespace = getattr(args, 'namespace', None)
    release_name = 'ob-demo' if release_name is None else release_name
    namespace = 'ob-demo' if namespace is None else namespace
    if len(namespace) > 63 or not re.fullmatch(r'[a-z0-9](?:[a-z0-9-]*[a-z0-9])?', namespace):
        raise ValueError('Invalid Helm namespace name')
    chart = one_chart(bundle, 'oceanbase-lakehouse-*.tgz')
    snapshots = [path.read_bytes() for path in args.values]
    endpoint, skip_operator = helm_values(snapshots)
    skip_crds = getattr(args, 'skip_crds', False) or skip_operator
    with tempfile.TemporaryDirectory(prefix='lakehouse-values-') as tmp:
        cmd = [args.helm, 'template', release_name, chart, '--namespace', namespace]
        for index, content in enumerate(snapshots):
            snapshot = Path(tmp) / ('values-%d.yaml' % index)
            write_private(snapshot, content)
            cmd += ['--values', snapshot]
        service = SERVICE_ENDPOINT.fullmatch(endpoint) if isinstance(endpoint, str) else None
        if service and (service[1] is None or 0 < int(service[1]) <= 65535):
            # Only the image plan uses this reserved test address. Keep the
            # original endpoint for install, where Helm validates the Service
            # and any retained storage identity against the target cluster.
            cmd += ['--set-string', 'objectStorage.endpoint=http://192.0.2.1']
        rendered = run(cmd)
    if not release or list(bundle.glob('ob-operator-*.tgz')):
        operator = [args.helm, 'template', 'ob-operator', one_chart(bundle, 'ob-operator-*.tgz'),
                    '--namespace', 'oceanbase-system']
        if (bundle / 'operator-values.yaml').exists():
            operator += ['--values', bundle / 'operator-values.yaml']
        rendered += '\n---\n' + run(operator)
    documents = list(yaml.safe_load_all(rendered))
    refs = manifest_refs(documents, excluded)
    bundled_prerequisites = any(
        doc and doc.get('kind') == 'Job' and doc['metadata']['name'].startswith('lakehouse-cert-')
        for doc in documents)
    administrator_assets = {}
    if getattr(args, 'include_prerequisites', False):
        cert_manager = bundle / 'cert-manager.yaml'
        if cert_manager.is_file():
            administrator_assets['cert_manager'] = cert_manager.name
            # Explicit administrator manifests include the dynamic HTTP01
            # solver in their lock. Older bundled self-signed setups did not.
            refs.update(manifest_refs(yaml.safe_load_all(cert_manager.read_text()), set(), include_solver=True))
        with tarfile.open(chart) as archive:
            payload_dir = 'oceanbase-lakehouse/files/prerequisites/'
            if payload_dir + 'bootstrap.py' in archive.getnames():
                archive.getmember(payload_dir + 'prepare.sh')
                payload = archive.extractfile(payload_dir + 'operator.yaml')
                with payload:
                    refs.update(manifest_refs(yaml.safe_load_all(payload.read()), set()))
                administrator_assets['operator'] = payload_dir + 'prepare.sh'
        if not administrator_assets:
            raise ValueError('This release has no administrator prerequisite payload')
    return refs, snapshots, {'release': release_name, 'namespace': namespace, 'skip_crds': skip_crds,
                              'bundled_prerequisites': bundled_prerequisites,
                              'administrator_assets': administrator_assets}


def write_private(path, data):
    with path.open('xb') as output:
        os.fchmod(output.fileno(), 0o600)
        output.write(data)


def plan(args):
    release = None
    if getattr(args, 'release_manifest', None):
        bundle, lock, release = manifest_bundle(args)
    else:
        bundle = args.bundle.resolve()
        lock = json.loads((bundle / 'image-lock.json').read_text())
    channel = lock['delivery_channel']
    expected = getattr(args, 'expected_channel', None)
    if expected and channel != expected:
        raise ValueError('This %s entry point cannot process a %s bundle' % (expected, channel))
    if channel == 'compose':
        if args.values or any(getattr(args, option, None) for option in
                              ('release', 'namespace', 'skip_crds', 'include_prerequisites')):
            raise ValueError('Helm values and installation options apply only to Helm')
        env_file = getattr(args, 'env_file', None)
        if env_file is None and not release and (bundle / '.env').is_file():
            env_file = bundle / '.env'
        command = [args.docker, 'compose', '--project-directory', bundle]
        if env_file is not None:
            command += ['--env-file', Path(env_file).resolve()]
        rendered = json.loads(run(command + ['-f', bundle / 'compose.yaml',
                                   # Keep generated network/volume names project-scoped.
                                   # Normalizing them here would bake in the exporter's
                                   # project name and share data across later -p installs.
                                   'config', '--no-path-resolution', '--no-normalize',
                                   '--format', 'json']))
        if any(s.get('build') for s in rendered['services'].values()):
            raise ValueError('Offline delivery requires prebuilt images, not Compose build directives')
        refs = {s['image'] for s in rendered['services'].values()}
    elif channel == 'helm':
        if getattr(args, 'env_file', None):
            raise ValueError('--env-file applies to Compose only; use --values for Helm')
        excluded = {v['reference'] for v in lock.get('excluded_optional', {}).values()}
        refs, snapshots, helm_install = helm_image_plan(args, bundle, release, excluded)
    else:
        raise ValueError('Unknown delivery channel: ' + channel)
    known = {i['reference']: i for i in lock['images']}
    unknown = refs - known.keys()
    if unknown:
        raise ValueError('Images outside this release lock: ' + ', '.join(sorted(unknown)) +
                         '. Supply a reviewed matching lock; images are never silently replaced.')
    if not refs:
        raise ValueError('No runtime images found')
    selected = {'channel': channel, 'platform': lock['platform'],
            'delivery_version': lock['delivery_version'],
            'prerequisites': lock.get('prerequisites', []) if channel == 'helm' else ['Docker Engine and Compose v2'],
            'images': [known[r] for r in sorted(refs)]}
    if channel == 'compose':
        selected['_compose_config'] = rendered
    else:
        selected['_helm_values'] = snapshots
        selected['helm_installation'] = helm_install
    if release:
        selected['release_manifest'] = release
        selected['_release_assets'] = list(args._release_assets)
    return selected


def saved_docker_config_matches(docker, reference, config_id, platform):
    """Read the config from the local Engine without writing an export to disk."""
    digest = config_id.split(':', 1)[1]
    config_paths = {digest + '.json', 'blobs/sha256/' + digest}
    verified = set()
    manifest = []
    command = [docker, 'image', 'save', '--platform', platform, reference]
    with subprocess.Popen(command, stdout=subprocess.PIPE) as process:
        try:
            with tarfile.open(fileobj=process.stdout, mode='r|*') as archive:
                for member in archive:
                    if member.name == 'manifest.json':
                        manifest = json.load(archive.extractfile(member))
                    elif member.name in config_paths:
                        config = archive.extractfile(member).read()
                        if 'sha256:' + hashlib.sha256(config).hexdigest() == config_id:
                            verified.add(member.name)
        except (tarfile.TarError, ValueError) as exc:
            raise ValueError('Cannot inspect local Docker image export: ' + reference) from exc
        finally:
            process.stdout.close()
        if process.wait():
            raise ValueError('Cannot export local Docker image: ' + reference)
    return len(manifest) == 1 and manifest[0].get('Config') in verified


def inspect_image(docker, item, platform):
    reference = item['reference']
    actual, = json.loads(run([docker, 'image', 'inspect', reference]))
    if actual['Os'] + '/' + actual['Architecture'] != platform:
        raise ValueError('Image platform mismatch: ' + reference)
    if actual['Id'] != item['image_id']:
        # The containerd image store reports an OCI descriptor instead of a config ID.
        descriptor = actual.get('Descriptor') or {}
        if actual['Id'] != descriptor.get('digest') or not saved_docker_config_matches(
                docker, reference, item['image_id'], platform):
            raise ValueError('Image identity mismatch: ' + reference)
    return actual


def sha256(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(8 * 1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def download(args, selected):
    metadata = []
    for item in selected['images']:
        ref = item['reference']
        if args.pull == 'always':
            subprocess.run([args.docker, 'pull', '--platform', selected['platform'], ref], check=True)
        elif args.pull == 'missing':
            exists = subprocess.run([args.docker, 'image', 'inspect', ref],
                                    stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0
            if not exists:
                subprocess.run([args.docker, 'pull', '--platform', selected['platform'], ref], check=True)
        # An existing but wrong image is an error, never a fallback to latest.
        actual = inspect_image(args.docker, item, selected['platform'])
        metadata.append({'reference': ref, 'image_id': item['image_id'],
                         'repo_digests': actual.get('RepoDigests', [])})
    return metadata


def offline_references(selected):
    """Docker archives need tags: docker load need not restore RepoDigests."""
    if selected['channel'] != 'compose':
        return {}
    if not selected.get('_compose_config'):
        if any('@' in item['reference'] for item in selected['images']):
            raise ValueError('Digest-pinned Compose offline delivery requires a rendered configuration for explicit local aliases')
        return {}
    if not selected.get('release_manifest') and not any('@' in i['reference'] for i in selected['images']):
        return {}
    result = {}
    for item in selected['images']:
        identity = item['image_id']
        if not identity.startswith('sha256:') or len(identity) != 71:
            raise ValueError('Offline aliases require a locked SHA256 config ID')
        suffix = hashlib.sha256(item['reference'].encode()).hexdigest()[:12]
        # A namespace-only alias is normalized to docker.io/... by skopeo's
        # Docker archive transport. Use an explicit registry component so the
        # requested reference and archive RepoTags stay exactly identical. This
        # is a local cache alias, never a registry contact (pull_policy=never).
        result[item['reference']] = 'localhost/lakehouse-offline/compose-' + suffix + ':' + identity[7:]
    return result


def offline_compose(selected, references):
    document = copy.deepcopy(selected['_compose_config'])
    for service in document['services'].values():
        if service.get('image') in references:
            service['image'] = references[service['image']]
            service['pull_policy'] = 'never'
    return yaml.safe_dump(document, sort_keys=False)


def tag_verified_images(args, selected, references):
    for item in selected['images']:
        alias = references.get(item['reference'])
        if not alias:
            continue
        present = subprocess.run([args.docker, 'image', 'inspect', alias],
                                 stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0
        if present:
            inspect_image(args.docker, {**item, 'reference': alias}, selected['platform'])
        else:
            subprocess.run([args.docker, 'image', 'tag', item['reference'], alias], check=True)


def export(args, selected):
    output = args.output.resolve()
    bundle = args.bundle.resolve()
    if output == bundle or bundle in output.parents:
        raise ValueError('--output must be outside the source bundle')
    if selected.get('release_manifest'):
        raise ValueError('Release-manifest mode uses --tar; --output remains available for legacy --bundle')
    output.mkdir(parents=True, exist_ok=False, mode=0o700)
    (output / 'INCOMPLETE').write_text('Do not use until packaging completes successfully.\n')
    metadata = download(args, selected)
    references = offline_references(selected)
    tag_verified_images(args, selected, references)
    target = output / 'deployment'
    shutil.copytree(bundle, target, ignore=shutil.ignore_patterns('__pycache__', '*.pyc'))
    if selected['channel'] == 'compose':
        # Freeze interpolated configuration while retaining relative bind mounts.
        # Compose config already escapes literal dollar signs for round trips.
        rendered = offline_compose(selected, references) if references else yaml.safe_dump(selected['_compose_config'], sort_keys=False)
        (target / 'compose.yaml').write_text(rendered)
    value_args = []
    if args.values:
        inputs = target / 'offline-inputs'
        inputs.mkdir()
        for n, path in enumerate(args.values):
            name = 'values-%d.yaml' % n
            shutil.copyfile(path, inputs / name)
            value_args += ['-f', 'offline-inputs/' + name]
    deployment_checks = [sha256(p) + '  ' + p.relative_to(target).as_posix()
                         for p in sorted(target.rglob('*')) if p.is_file() and p != target / 'SHA256SUMS']
    (target / 'SHA256SUMS').write_text('\n'.join(deployment_checks) + '\n')
    # The archive is a single Docker save stream, so shared layers are deduplicated.
    partial = output / 'images.tar.gz.partial'
    with partial.open('wb') as raw, gzip.GzipFile(fileobj=raw, mode='wb', mtime=0) as compressed:
        process = subprocess.Popen([args.docker, 'image', 'save'] +
                                   [references.get(i['reference'], i['reference']) for i in selected['images']], stdout=subprocess.PIPE)
        try:
            shutil.copyfileobj(process.stdout, compressed, 1024 * 1024)
        finally:
            process.stdout.close()
        if process.wait():
            raise ValueError('docker image save failed; output is incomplete')
    partial.rename(output / 'images.tar.gz')
    selected.update(inspected_images=metadata, deployment_tested=False)
    public = {k: v for k, v in selected.items() if not k.startswith('_')}
    if references:
        public['offline_references'] = references
    (output / 'offline-manifest.json').write_text(json.dumps(public, indent=2) + '\n')
    if selected['channel'] == 'compose':
        commands = 'gzip -dc images.tar.gz | docker load\ncd deployment\ndocker compose up -d --pull never\n'
    else:
        chart = one_chart(target, 'oceanbase-lakehouse-*.tgz').name
        commands = ('# Import on EVERY target K3s worker (Docker load does not populate K3s).\n'
                    'gzip -dc images.tar.gz | sudo k3s ctr images import -\n'
                    'cd deployment\n# Check the Kubernetes and storage prerequisites in USAGE.txt.\n'
                    'helm upgrade --install ob-demo ./' + chart + ' -n ob-demo --create-namespace ' +
                    ' '.join(value_args) + ' --wait --wait-for-jobs --timeout 35m\n')
    (output / 'INSTALL.txt').write_text('Verify first: sha256sum -c SHA256SUMS\n\n' + commands +
                                      '\nNo cluster installation was performed by the packaging tool.\n')
    checks = [sha256(p) + '  ' + p.relative_to(output).as_posix()
              for p in sorted(output.rglob('*')) if p.is_file() and p.name != 'INCOMPLETE']
    (output / 'SHA256SUMS').write_text('\n'.join(checks) + '\n')
    (output / 'INCOMPLETE').unlink()
    return output


def oci_archive(args, item, platform, target):
    """Keep the registry's original manifest and compressed blobs."""
    reference = item['reference']
    digest = reference.rsplit('@', 1)[1]
    if not SHA256.fullmatch(digest) or digest != item.get('manifest_digest', digest):
        raise ValueError('Invalid locked registry manifest for ' + reference)
    os_name, arch = platform.split('/')
    with tempfile.TemporaryDirectory(prefix='.registry-image-', dir=target.parent) as tmp:
        directory = Path(tmp)
        # The oci transport may convert Docker schema-2 manifests to OCI.
        # The dir transport preserves the source bytes regardless of media type.
        subprocess.run([args.skopeo, 'copy', '--all', '--preserve-digests',
                        'docker://' + reference, 'dir:' + str(directory)], check=True)
        original = directory / 'manifest.json'
        if 'sha256:' + sha256(original) != digest:
            raise ValueError('Registry manifest digest mismatch: ' + reference)
        blobs = {digest: original}
        platform_configs = set()

        def blob(descriptor, suffix=''):
            identity = descriptor['digest']
            if not isinstance(identity, str) or not SHA256.fullmatch(identity):
                raise ValueError('Invalid OCI blob digest for ' + reference)
            path = blobs.get(identity)
            if path is None:
                path = directory / (identity[7:] + suffix)
                if 'sha256:' + sha256(path) != identity:
                    raise ValueError('OCI blob digest mismatch: ' + identity)
                blobs[identity] = path
            if path.stat().st_size != descriptor['size']:
                raise ValueError('OCI blob size mismatch: ' + identity)
            return path

        def visit(path):
            manifest = json.loads(path.read_bytes())
            if manifest.get('schemaVersion') != 2:
                raise ValueError('Unsupported registry manifest schema: ' + reference)
            if 'manifests' in manifest:
                for child in manifest['manifests']:
                    visit(blob(child, '.manifest.json'))
                return manifest.get('mediaType', 'application/vnd.oci.image.index.v1+json')
            config = manifest['config']
            config_data = json.loads(blob(config).read_bytes())
            if config_data.get('os') == os_name and config_data.get('architecture') == arch:
                platform_configs.add(config['digest'])
            for layer in manifest['layers']:
                blob(layer)
            if config.get('mediaType') == 'application/vnd.docker.container.image.v1+json':
                default_type = 'application/vnd.docker.distribution.manifest.v2+json'
            else:
                default_type = 'application/vnd.oci.image.manifest.v1+json'
            return manifest.get('mediaType', default_type)

        media_type = visit(original)
        if platform_configs != {item['image_id']}:
            raise ValueError('Image identity/platform mismatch or ambiguity: ' + reference)
        index = {'schemaVersion': 2, 'manifests': [{
            'mediaType': media_type, 'digest': digest, 'size': original.stat().st_size,
            'annotations': {'io.containerd.image.name': reference,
                            'org.opencontainers.image.ref.name': reference},
        }]}
        with tarfile.open(target, 'w') as archive:
            for name, content in [('oci-layout', {'imageLayoutVersion': '1.0.0'}),
                                  ('index.json', index)]:
                data = json.dumps(content).encode()
                member = tarfile.TarInfo(name)
                member.size = len(data)
                archive.addfile(member, io.BytesIO(data))
            for identity, path in sorted(blobs.items()):
                archive.add(path, arcname='blobs/sha256/' + identity[7:], recursive=False)


CRI_IMAGE_FORMAT = '{{.status.id}}|{{.info.imageSpec.os}}/{{.info.imageSpec.architecture}}'
K3S_CONTAINERD_SOCKET = '/run/k3s/containerd/containerd.sock'
K3S_CRI_IMAGE_OPTIONS = ['crictl', '--config', '/dev/null', '--image-endpoint']


def k3s_identity_check(item, platform):
    args = ['inspecti', '--output', 'go-template', '--template',
            CRI_IMAGE_FORMAT, item['reference']]
    return args, item['image_id'] + '|' + platform


def digest_lookup(item):
    reference = item['reference']
    args = ['ctr', '-n', 'k8s.io', 'images', 'ls', 'name==' + json.dumps(reference)]
    return args, reference, reference.rsplit('@', 1)[1]


def k3s_verification_script(images, platform):
    lines = ['cri_endpoint="unix://${CONTAINERD_ADDRESS:-' + K3S_CONTAINERD_SOCKET + '}"']
    for item in images:
        command, wanted = k3s_identity_check(item, platform)
        lines.extend([
            'actual=$("${K3S:-k3s}" ' + shlex.join(K3S_CRI_IMAGE_OPTIONS) +
            ' "$cri_endpoint" ' + shlex.join(command) + ')',
            '[ "$actual" = ' + shlex.quote(wanted) +
            ' ] || { echo "Loaded K3s image identity/platform mismatch" >&2; exit 1; }',
        ])
        if '@' in item['reference']:
            command, ref, digest = digest_lookup(item)
            lines.extend([
                'records=$("${K3S:-k3s}" ' + shlex.join(command) + ')',
                'printf "%s\\n" "$records" | awk -v ref=' + shlex.quote(ref) +
                ' -v digest=' + shlex.quote(digest) +
                " '$1 == ref && $3 == digest { found=1 } END { exit !found }'" +
                ' || { echo "Loaded K3s reference/manifest digest mismatch" >&2; exit 1; }',
            ])
    return '\n'.join(lines) + '\n'


def helm_installation_text(selected, value_args):
    settings = selected['helm_installation']
    administrator = settings['administrator_assets']
    instructions = ('Import on EVERY target K3s worker. Docker load does not populate K3s.\n'
                    '  sh load.sh\n'
                    '  cd deployment\n')
    if settings.get('bundled_prerequisites'):
        instructions += '\nThe Helm command below installs the bundled shared components.\n'
    elif administrator:
        instructions += '\nFor a new cluster, an administrator first prepares shared prerequisites.\n'
        if 'cert_manager' in administrator:
            instructions += ('  kubectl apply --server-side -f ' + administrator['cert_manager'] + '\n'
                             '  kubectl -n cert-manager wait --for=condition=Available '
                             'deployment/cert-manager deployment/cert-manager-cainjector '
                             'deployment/cert-manager-webhook --timeout=300s\n')
        if 'operator' in administrator:
            instructions += ('  tar -xzf oceanbase-lakehouse-release.tgz\n'
                             '  sh ' + administrator['operator'] + '\n')
    else:
        instructions += '\nPrepare any administrator-managed prerequisites as described by this release.\n'
    command = ['helm', 'upgrade', '--install', settings['release'], './oceanbase-lakehouse-release.tgz',
               '-n', settings['namespace'], '--create-namespace']
    if settings['skip_crds']:
        command.append('--skip-crds')
    command += value_args + ['--wait', '--wait-for-jobs', '--timeout', '35m']
    return (instructions + '\nDeploy using the planned release, namespace and values.\n'
            '  ' + shlex.join(command) + '\n')


def archive_images(args, selected):
    """Download registry images without contacting a local Docker daemon."""
    if args.tar == Path('__AUTO__'):
        version = str(selected['delivery_version'])
        if not version or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-' for c in version):
            raise ValueError('Invalid delivery version for archive filename')
        output = Path.cwd() / ('oceanbase-lakehouse-' + selected['channel'] + '-images-' +
                               version + '-' + selected['platform'].replace('/', '-') + '.tar')
    else:
        output = args.tar.resolve()
    if output.exists():
        raise ValueError('--tar destination already exists: ' + str(output))
    if args.pull == 'never':
        raise ValueError('--tar downloads directly from the registry; --pull never is unsupported')
    references = offline_references(selected)
    output.parent.mkdir(parents=True, exist_ok=True)
    os_name, arch = selected['platform'].split('/')
    # Stage beside the destination; failed downloads never produce a final tar.
    with tempfile.TemporaryDirectory(prefix='.offline-', dir=output.parent) as tmp:
        root = Path(tmp)
        images = root / 'images'
        images.mkdir()
        for n, item in enumerate(selected['images']):
            path = images / ('%03d.tar' % n)
            if selected['channel'] == 'helm' and '@' in item['reference']:
                oci_archive(args, item, selected['platform'], path)
                continue
            archive_ref = references.get(item['reference'], item['reference'])
            subprocess.run([args.skopeo, 'copy', '--override-os', os_name,
                            '--override-arch', arch, 'docker://' + item['reference'],
                            'docker-archive:' + str(path) + ':' + archive_ref], check=True)
            with tarfile.open(path) as image:
                manifest = json.load(image.extractfile('manifest.json'))
                if len(manifest) != 1 or manifest[0].get('RepoTags') != [archive_ref]:
                    raise ValueError('Unexpected archive image: ' + item['reference'])
                config = image.extractfile(manifest[0]['Config']).read()
                if 'sha256:' + hashlib.sha256(config).hexdigest() != item['image_id']:
                    raise ValueError('Image identity mismatch: ' + item['reference'])
                config_data = json.loads(config)
                if config_data.get('os') != os_name or config_data.get('architecture') != arch:
                    raise ValueError('Image platform mismatch: ' + item['reference'])
        public = {k: v for k, v in selected.items() if not k.startswith('_')}
        if references:
            public['offline_references'] = references
        (root / 'offline-manifest.json').write_text(json.dumps(public, indent=2) + '\n')
        additions = []
        installation = ''
        if references:
            (root / 'compose.offline.yaml').write_text(offline_compose(selected, references))
            (root / 'compose.offline.yaml').chmod(0o600)
            additions.append('compose.offline.yaml')
            installation = ('Docker load does not guarantee registry RepoDigests. This configuration explicitly maps\n'
                            'each original image to a local tag verified against its locked config ID/platform.\n'
                            'Every service uses pull_policy: never.\n'
                            'The original online @digest configuration is unchanged.\n\n'
                            '  sh load.sh &&\n'
                            '  docker compose -f compose.offline.yaml up -d --pull never\n'
                            '  docker compose -f compose.offline.yaml logs -f observer-bootstrap\n\n'
                            'Read the password locally using the same Compose prefix as startup:\n'
                            '  docker compose -f compose.offline.yaml run --rm --no-deps observer-bootstrap --mode show-password\n\n'
                            'For an existing instance include its original -p PROJECT_NAME in each\n'
                            'Compose command and retain the volumes and configuration.\n')
        elif selected.get('_release_assets'):
            target = root / 'deployment'
            target.mkdir()
            for asset in selected['_release_assets']:
                shutil.copy2(asset, target / asset.name)
            value_args = []
            for n, value in enumerate(selected.get('_helm_values', [])):
                filename = 'values-%d.yaml' % n
                write_private(target / filename, value)
                value_args += ['-f', filename]
            additions.append('deployment')
            installation = helm_installation_text(selected, value_args)
        if installation:
            (root / 'INSTALL.txt').write_text('This is a business-image/configuration attachment, not an OS/runtime installer.\n'
                                              'Extract in a protected empty directory; it may contain expanded user configuration.\n'
                                              'The loader verifies archive checksums and image config IDs/platforms.\n\n' + installation)
            (root / 'README.md').write_text('# Offline deployment attachment\n\n' + (root / 'INSTALL.txt').read_text())
            additions += ['INSTALL.txt', 'README.md']
        loader = root / 'load.sh'
        import_command = ('for image in images/*.tar; do "${DOCKER:-docker}" load -i "$image"; done\n'
                          if selected['channel'] == 'compose' else
                          'for image in images/*.tar; do "${K3S:-k3s}" ctr -n k8s.io images import --platform ' +
                          shlex.quote(selected['platform']) + ' "$image"; done\n')
        verification = ''
        if selected['channel'] == 'compose':
            verifier = root / 'verify-docker-image.py'
            verifier.write_text(
                'import hashlib\nimport json\nimport subprocess\nimport sys\nimport tarfile\n\n'
                + '\n\n'.join(inspect.getsource(function) for function in
                              (run, saved_docker_config_matches, inspect_image))
                + '\ndocker, reference, config_id, platform = sys.argv[1:]\n'
                  'try:\n'
                  '    inspect_image(docker, {"reference": reference, "image_id": config_id}, platform)\n'
                  'except ValueError as error:\n'
                  '    raise SystemExit(str(error))\n')
            additions.append(verifier.name)
            for item in selected['images']:
                ref = references.get(item['reference'], item['reference'])
                wanted = item['image_id'] + '|' + selected['platform']
                verification += ('actual=$("${DOCKER:-docker}" image inspect --format '
                                 + shlex.quote('{{.Id}}|{{.Os}}/{{.Architecture}}') + ' ' + shlex.quote(ref) + ')\n'
                                 'if [ "$actual" != ' + shlex.quote(wanted) + ' ]; then\n'
                                 '  command -v python3 >/dev/null || { echo "Docker config verification needs Python 3" >&2; exit 1; }\n'
                                 '  python3 verify-docker-image.py "${DOCKER:-docker}" '
                                 + shlex.join([ref, item['image_id'], selected['platform']]) + '\nfi\n')
        else:
            verification = k3s_verification_script(selected['images'], selected['platform'])
        loader.write_text('#!/bin/sh\nset -eu\ncd "$(dirname "$0")"\n'
                          'if command -v sha256sum >/dev/null 2>&1; then sha256sum -c SHA256SUMS; else shasum -a 256 -c SHA256SUMS; fi\n'
                          + import_command + verification)
        loader.chmod(0o755)
        files = sorted(p for p in root.rglob('*') if p.is_file())
        (root / 'SHA256SUMS').write_text(''.join(
            sha256(p) + '  ' + p.relative_to(root).as_posix() + '\n' for p in files))
        staging = root / 'bundle.tar'
        with tarfile.open(staging, 'w') as tar:
            for name in ['images', 'offline-manifest.json', 'load.sh', 'SHA256SUMS'] + additions:
                tar.add(root / name, arcname=name)
        # The outer archive can contain expanded user credentials. Protect the
        # archive itself as well as its individual configuration members before
        # publishing it outside the private temporary directory.
        staging.chmod(0o600)
        # Exclusive creation protects an existing file even if it appeared during download.
        try:
            os.link(staging, output)
        except FileExistsError:
            raise ValueError('--tar destination already exists: ' + str(output))
    return output


def download_k3s(args, selected):
    """Validate registry archives before importing into this K3s node only."""
    if args.pull == 'never':
        raise ValueError('K3s mode downloads from the registry; --pull never is unsupported')
    # Fail before downloading if the runtime is unavailable or access is denied.
    run([args.k3s, 'ctr', '-n', 'k8s.io', 'version'])
    with tempfile.TemporaryDirectory(prefix='lakehouse-k3s-') as tmp:
        options = argparse.Namespace(**vars(args))
        options.tar = Path(tmp) / 'images.tar'
        archive_images(options, selected)
        with tarfile.open(options.tar) as archive:
            # This archive was generated and validated by this process. Extract
            # only image streams, never arbitrary paths or executable content.
            images = sorted(m for m in archive.getnames()
                            if m.startswith('images/') and m.endswith('.tar'))
            for n, name in enumerate(images):
                path = Path(tmp) / ('image-%03d.tar' % n)
                with archive.extractfile(name) as src, path.open('wb') as dest:
                    shutil.copyfileobj(src, dest)
                subprocess.run([args.k3s, 'ctr', '-n', 'k8s.io', 'images', 'import',
                                '--platform', selected['platform'], str(path)], check=True)
                path.unlink()
    endpoint = 'unix://' + (os.environ.get('CONTAINERD_ADDRESS') or K3S_CONTAINERD_SOCKET)
    for item in selected['images']:
        command, wanted = k3s_identity_check(item, selected['platform'])
        if run([args.k3s] + K3S_CRI_IMAGE_OPTIONS + [endpoint] + command).strip() != wanted:
            raise ValueError('Loaded K3s image identity/platform mismatch: ' + item['reference'])
        if '@' in item['reference']:
            command, ref, digest = digest_lookup(item)
            fields = [line.split() for line in run([args.k3s] + command).splitlines()]
            if not any(len(row) >= 3 and row[0] == ref and row[2] == digest for row in fields):
                raise ValueError('Loaded K3s reference/manifest digest mismatch: ' + ref)
    print('Images imported into this K3s node (namespace k8s.io). Repeat on every schedulable node.')


def main(expected_channel=None, default_manifest=None, default_manifest_sha256=None):
    parser = argparse.ArgumentParser(description=__doc__)
    source = parser.add_mutually_exclusive_group()
    source.add_argument('--bundle', type=Path, help='Legacy extracted delivery directory')
    source.add_argument('--release-manifest', help='Versioned release-manifest.json URL or local path; downloads and verifies configuration/Charts, not images in --list mode')
    parser.add_argument('--manifest-sha256', help='Optional trusted SHA256 of release-manifest.json')
    parser.add_argument('--channel', choices=['compose', 'helm'], help='Compatibility entry only; fixed-channel tools do not require this')
    parser.add_argument('--env-file', type=Path, help='Optional user Compose interpolation file; no .env is required in release-manifest mode')
    parser.add_argument('--values', type=Path, action='append', default=[])
    parser.add_argument('--release', help='Helm release used for planning and the offline install command (default ob-demo)')
    parser.add_argument('--namespace', help='Helm namespace used for planning and the offline install command (default ob-demo)')
    parser.add_argument('--skip-crds', action='store_true', help='Include --skip-crds in the Helm offline install command')
    parser.add_argument('--include-prerequisites', action='store_true',
                        help='Include the release administrator bootstrap images and available cert-manager manifest')
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument('--output', type=Path, help='Create an offline directory; must not already exist')
    mode.add_argument('--tar', type=Path, nargs='?', const=Path('__AUTO__'), help='Create a tar without local Docker import; defaults to oceanbase-lakehouse-CHANNEL-images-VERSION-PLATFORM.tar in the current directory (requires skopeo)')
    mode.add_argument('--list', action='store_true', help='List images without downloading')
    mode.add_argument('--check', action='store_true', help='Check local Docker image IDs without pulling')
    parser.add_argument('--pull', choices=['always', 'missing', 'never'], default='missing')
    parser.add_argument('--docker', default='docker')
    parser.add_argument('--helm', default='helm')
    parser.add_argument('--skopeo', default='skopeo')
    parser.add_argument('--k3s', default='k3s', help='K3s executable (Helm entry point only)')
    args = parser.parse_args()
    if expected_channel and args.channel and args.channel != expected_channel:
        parser.error('This tool is fixed to the ' + expected_channel + ' channel')
    args.expected_channel = expected_channel or args.channel
    if args.bundle is None and args.release_manifest is None:
        args.release_manifest = default_manifest
        if args.manifest_sha256 is None:
            args.manifest_sha256 = default_manifest_sha256
        if args.release_manifest is None:
            args.bundle = Path(__file__).resolve().parent.parent
    try:
        selected = plan(args)
        print(json.dumps({k: v for k, v in selected.items() if not k.startswith('_')}, indent=2), flush=True)
        if selected['channel'] == 'helm' and (args.output or args.check):
            raise ValueError('Helm entry does not use Docker cache: use --tar, --list, or default K3s import')
        if args.output:
            print('Offline package ready: ' + str(export(args, selected)))
        elif args.tar:
            print('Image tar ready: ' + str(archive_images(args, selected)))
        elif args.check:
            errors = []
            for item in selected['images']:
                try:
                    inspect_image(args.docker, item, selected['platform'])
                except ValueError as error:
                    errors.append(str(error))
            if errors:
                raise ValueError('\n'.join(errors))
            print('All selected images match the release lock.')
        elif not args.list:
            if selected['channel'] == 'helm':
                download_k3s(args, selected)
            else:
                download(args, selected)
                print('All selected images are loaded in the current Docker environment.')
    except (ValueError, OSError, KeyError, yaml.YAMLError, tarfile.TarError, subprocess.SubprocessError) as error:
        parser.exit(1, 'Offline packaging failed: %s\n' % error)
    finally:
        context = getattr(args, '_release_context', None)
        if context is not None:
            context.cleanup()


if __name__ == '__main__':
    main()

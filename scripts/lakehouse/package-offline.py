#!/usr/bin/env python3
"""List, check, or export the exact images used by a native delivery bundle.

Python 3.8+, PyYAML, Docker Compose (Compose bundles), Helm (Helm bundles).
No deployment, image substitution, registry push, or cluster modification.
"""
import argparse
import copy
import tarfile
import gzip
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

import yaml


def run(argv):
    result = subprocess.run([str(x) for x in argv], stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE, text=True, timeout=300)
    if result.returncode:
        raise ValueError('Command failed: %s\n%s' %
                         (' '.join(map(str, argv)), result.stderr.strip()))
    return result.stdout


def refs_in(value):
    if isinstance(value, dict):
        for key, child in value.items():
            if key == 'image' and isinstance(child, str):
                yield child
            else:
                yield from refs_in(child)
    elif isinstance(value, list):
        for child in value:
            yield from refs_in(child)


def manifest_refs(documents, excluded):
    refs = set()
    for doc in documents:
        if not isinstance(doc, dict):
            continue
        # Includes OBCluster/OBLogServiceCluster and topology overrides, not
        # only Pod images. These CR images are instantiated later by Operator.
        refs.update(refs_in(doc))
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
                refs.update(set(refs_in(embedded)) - excluded)
    return refs


def one_chart(bundle, pattern):
    matches = list(bundle.glob(pattern))
    if len(matches) != 1:
        raise ValueError('Expected exactly one ' + pattern)
    return matches[0]


def plan(args):
    bundle = args.bundle.resolve()
    lock = json.loads((bundle / 'image-lock.json').read_text())
    channel = lock['delivery_channel']
    if channel == 'compose':
        if args.values:
            raise ValueError('--values is only supported for Helm bundles; edit Compose .env instead')
        rendered = json.loads(run([args.docker, 'compose', '--project-directory', bundle,
                                   '--env-file', (bundle / '.env' if (bundle / '.env').exists() else Path('/dev/null')), '-f', bundle / 'compose.yaml',
                                   # Keep generated network/volume names project-scoped.
                                   # Normalizing them here would bake in the exporter's
                                   # project name and share data across later -p installs.
                                   'config', '--no-path-resolution', '--no-normalize',
                                   '--format', 'json']))
        if any(s.get('build') for s in rendered['services'].values()):
            raise ValueError('Offline delivery requires prebuilt images, not Compose build directives')
        refs = {s['image'] for s in rendered['services'].values()}
    elif channel == 'helm':
        cmd = [args.helm, 'template', 'ob-demo', one_chart(bundle, 'oceanbase-lakehouse-*.tgz'),
               '--namespace', 'ob-demo']
        for path in args.values:
            cmd += ['--values', path.resolve()]
        rendered = run(cmd)
        rendered += '\n---\n' + run([args.helm, 'template', 'ob-operator',
                                    one_chart(bundle, 'ob-operator-*.tgz'),
                                    '--namespace', 'oceanbase-system',
                                    '--values', bundle / 'operator-values.yaml'])
        excluded = {v['reference'] for v in lock.get('excluded_optional', {}).values()}
        refs = manifest_refs(yaml.safe_load_all(rendered), excluded)
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


def offline_references(selected):
    """Docker archives need tags: docker load need not restore RepoDigests."""
    if selected["channel"] != "compose":
        return {}
    if selected['channel'] == 'compose' and not selected.get('_compose_config'):
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
        result[item['reference']] = 'localhost/lakehouse-offline/' + selected['channel'] + '-' + suffix + ':' + identity[7:]
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
    if selected.get('channel') == 'helm' and any('@' in item['reference'] for item in selected['images']):
        raise ValueError('Digest-pinned Helm images require OCI export: use package-offline-helm.py --tar with a release manifest')
    output = args.output.resolve()
    bundle = args.bundle.resolve()
    if output == bundle or bundle in output.parents:
        raise ValueError('--output must be outside the source bundle')
    output.mkdir(parents=True, exist_ok=False, mode=0o700)
    (output / 'INCOMPLETE').write_text('Do not use until packaging completes successfully.\n')
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
        metadata.append({'reference': ref, 'image_id': actual['Id'],
                         'repo_digests': actual.get('RepoDigests', [])})
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
    public = {key: value for key, value in selected.items() if not key.startswith('_')}
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
    loader = '#!/bin/sh\nset -eu\ncd -- "$(dirname -- "$0")"\nsha256sum -c SHA256SUMS\n'
    if selected['channel'] == 'compose':
        loader += 'gzip -dc images.tar.gz | "${DOCKER_BIN:-docker}" load\n'
    else:
        loader += 'gzip -dc images.tar.gz | "${K3S_BIN:-k3s}" ctr -n k8s.io images import -\n'
    (output / 'load.sh').write_text(loader)
    (output / 'load.sh').chmod(0o700)
    commands = commands.replace('gzip -dc images.tar.gz | docker load', 'sh load.sh').replace('gzip -dc images.tar.gz | sudo k3s ctr images import -', 'sudo sh load.sh')
    (output / 'INSTALL.txt').write_text('Verify first: sha256sum -c SHA256SUMS\n\n' + commands +
                                      '\nNo cluster installation was performed by the packaging tool.\n')
    checks = [sha256(p) + '  ' + p.relative_to(output).as_posix()
              for p in sorted(output.rglob('*')) if p.is_file() and p.name != 'INCOMPLETE']
    (output / 'SHA256SUMS').write_text('\n'.join(checks) + '\n')
    (output / 'INCOMPLETE').unlink()
    return output


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--bundle', type=Path, default=Path(__file__).resolve().parent.parent)
    parser.add_argument('--values', type=Path, action='append', default=[])
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument('--output', type=Path, help='Create an offline directory; must not already exist')
    mode.add_argument('--check', action='store_true', help='Check local Docker image IDs without pulling')
    parser.add_argument('--pull', choices=['always', 'missing', 'never'], default='missing')
    parser.add_argument('--docker', default='docker')
    parser.add_argument('--helm', default='helm')
    args = parser.parse_args()
    try:
        selected = plan(args)
        print(json.dumps({k: v for k, v in selected.items() if not k.startswith('_')}, indent=2), flush=True)
        if args.output:
            print('Offline package ready: ' + str(export(args, selected)))
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
        else:
            print('Plan only: no images pulled or saved. Use --output DIR to package.')
    except (ValueError, OSError, KeyError, yaml.YAMLError, subprocess.SubprocessError) as error:
        parser.exit(1, 'Offline packaging failed: %s\n' % error)


if __name__ == '__main__':
    main()

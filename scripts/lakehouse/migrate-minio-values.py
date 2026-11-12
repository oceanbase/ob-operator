#!/usr/bin/env python3
"""Migrate exported effective Helm values to the independent MinIO subchart.

Export with: helm get values RELEASE -n NAMESPACE --all -o yaml > previous.yaml
The tool reads local values only; it does not contact Kubernetes or change data.
"""
import argparse
import copy
import os
from pathlib import Path
import yaml


def migrate(values):
    if not isinstance(values, dict) or not isinstance(values.get('minio'), dict):
        raise ValueError('expected effective values from helm get values --all')
    result = copy.deepcopy(values)
    minio = result['minio']
    if 'enabled' in minio:
        return result  # Current values, including explicit empty overrides, are authoritative.
    storage = result.get('objectStorage', {})
    minio['enabled'] = storage.get('mode', 'bundled') != 'external'
    for key, source in {
        'storageClass': (result, 'storageClass'),
        'region': (storage, 'region'),
        'nodeSelector': (result.get('scheduling', {}), 'nodeSelector'),
        'tolerations': (result.get('scheduling', {}), 'tolerations'),
        'imagePullSecrets': (result, 'imagePullSecrets'),
    }.items():
        parent, field = source
        if key not in minio and field in parent:
            minio[key] = copy.deepcopy(parent[field])
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('input', type=Path)
    parser.add_argument('--output', required=True, type=Path)
    args = parser.parse_args()
    converted = migrate(yaml.safe_load(args.input.read_text()))
    # Effective values may contain user-provided passwords. Never overwrite an
    # existing file or expose the export through world-readable permissions.
    fd = os.open(args.output, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    try:
        with os.fdopen(fd, 'w') as stream:
            yaml.safe_dump(converted, stream, sort_keys=False)
    except BaseException:
        args.output.unlink(missing_ok=True)
        raise


if __name__ == '__main__':
    main()

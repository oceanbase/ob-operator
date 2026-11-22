#!/usr/bin/env python3
"""Print Helm image overrides for a source checkout of OceanSeek."""
import argparse
import json
from pathlib import Path
import re

import yaml


def version(value):
    if not re.fullmatch(r'[0-9]+\.[0-9]+\.[0-9]+', value):
        raise argparse.ArgumentTypeError('expected a component release version X.Y.Z')
    return value


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--registry', choices=('quay.io', 'ghcr.io', 'docker.io'), default='quay.io')
    parser.add_argument('--component-namespace', default='oceanbase',
                        help='publisher of Operator, Dashboard and Helper on the selected registry')
    for role in ('operator', 'dashboard', 'helper'):
        parser.add_argument('--' + role + '-version', required=True, type=version)
    args = parser.parse_args()
    if not re.fullmatch(r'[a-z0-9]+(?:[._-][a-z0-9]+)*', args.component_namespace):
        parser.error('expected a registry organization for --component-namespace')

    path = Path(__file__).resolve().parents[2] / 'charts/oceanseek/values.yaml'
    defaults = yaml.safe_load(path.read_text())
    overrides = {}
    for key in ('observer', 'logService', 'minio', 'initialization', 'monitor'):
        # The public registries carry the same product repository and release tag.
        repo = defaults[key]['image'].split('/', 1)[1]
        overrides[key] = {'image': args.registry + '/' + repo}

    publisher = args.registry + '/' + args.component_namespace
    overrides['native'] = {'readiness': {'image': publisher + '/ob-operator:' + args.operator_version}}
    helper = publisher + '/oceanbase-helper:' + args.helper_version
    overrides['prerequisites'] = {'image': helper}
    overrides['dashboard'] = {
        'image': publisher + '/oceanbase-dashboard:' + args.dashboard_version,
        'inspection': {'oceanbaseHelper': {'image': helper}},
    }
    print(json.dumps(overrides, indent=2))


if __name__ == '__main__':
    main()

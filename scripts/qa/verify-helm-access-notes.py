#!/usr/bin/env python3
"""Exercise packaged NOTES with real Helm lookup on a disposable test cluster.

Run only on an authorized test machine. Creates and deletes its own namespace;
no database workloads are needed for the address/licensing branch regression.
"""
import argparse
import copy
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile
import uuid
import yaml


def run(*args, input=None):
    p = subprocess.run(args, input=input, text=True, capture_output=True, check=True)
    return p.stdout


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('chart', type=Path)
    args = parser.parse_args()
    ns = 'notes-qa-' + uuid.uuid4().hex[:10]
    checks = []
    with tempfile.TemporaryDirectory() as tmp:
        chart = Path(tmp) / 'fixture'
        (chart / 'templates').mkdir(parents=True)
        with tarfile.open(args.chart) as tar:
            prefix = 'oceanbase-lakehouse/'
            for name in ('values.yaml', 'templates/NOTES.txt', 'templates/_helpers.tpl'):
                (chart / name).write_bytes(tar.extractfile(prefix + name).read())
        (chart / 'Chart.yaml').write_text('apiVersion: v2\nname: notes-fixture\nversion: 0.1.0\n')
        defaults = yaml.safe_load((chart / 'values.yaml').read_text())

        def render(name, host='qa.example.test', port=0, kind='NodePort', enabled=True,
                   endpoint='', minio=True):
            values = copy.deepcopy(defaults)
            values['access']['host'] = host
            values['dashboard']['enabled'] = enabled
            values['dashboard']['service'].update(type=kind, nodePort=port)
            values['objectStorage']['endpoint'] = endpoint
            values['minio']['enabled'] = minio
            (chart / 'values.yaml').write_text(yaml.safe_dump(values))
            output = run('helm', 'install', 'notes', str(chart), '-n', ns, '--dry-run=server')
            checks.append(name)
            return output.split('NOTES:', 1)[1]

        run('kubectl', 'create', 'namespace', ns)
        try:
            first = render('first-install-auto-port')
            assert 'not available while Helm renders' in first
            assert "NODE_HOST='qa.example.test'" in first
            assert 'get svc oceanbase-dashboard-notes' in first
            assert first.count('AGPL-3.0') == 1
            ipv6 = render('first-install-ipv6', host='2001:db8::1')
            assert "NODE_HOST='2001:db8::1'" in ipv6 and 'NODE_URL_HOST="[$NODE_HOST]"' in ipv6
            unsafe = render('host-shell-quoting', host="a'$(false)")
            assert "NODE_HOST='a'\"'\"'$(false)'" in unsafe
            explicit = render('explicit-port-ipv6', host='2001:db8::1', port=31183)
            assert 'Dashboard URL: http://[2001:db8::1]:31183/' in explicit
            service = {'apiVersion': 'v1', 'kind': 'Service',
                       'metadata': {'name': 'oceanbase-dashboard-notes', 'namespace': ns},
                       'spec': {'type': 'NodePort', 'ports': [{'port': 80, 'name': 'http'}]}}
            run('kubectl', 'apply', '-f', '-', input=json.dumps(service))
            actual = json.loads(run('kubectl', '-n', ns, 'get', 'svc',
                                    'oceanbase-dashboard-notes', '-o', 'json'))['spec']['ports'][0]['nodePort']
            existing = render('existing-service-auto-port')
            assert 'Dashboard URL: http://qa.example.test:%s/' % actual in existing
            assert 'Dashboard URL: http://qa.example.test:31184/' in render('explicit-overrides-existing', port=31184)
            assert 'port-forward svc/oceanbase-dashboard-notes' in render('cluster-ip', kind='ClusterIP')
            assert '<EXTERNAL-IP-or-hostname>' in render('load-balancer', kind='LoadBalancer')
            assert 'Dashboard is disabled' in render('dashboard-disabled', enabled=False)
            external = render('external-storage-no-minio', endpoint='https://s3.example.test', minio=False)
            assert external.count('AGPL-3.0') == 1 and 'no bundled MinIO server' in external
            separate = render('external-storage-with-minio', endpoint='https://s3.example.test')
            assert separate.count('AGPL-3.0') == 1 and 'separately enabled MinIO server' in separate
        finally:
            run('kubectl', 'delete', 'namespace', ns, '--wait=false')
    print(json.dumps({'status': 'PASS', 'checks': checks}, indent=2))


if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""Exercise packaged NOTES with real Helm lookup on a disposable test cluster.

Run only on an authorized test machine. Creates and deletes its own namespace;
no database workloads are needed for the access notes regression.
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
            prefix = 'oceanseek/'
            for name in ('values.yaml', 'templates/NOTES.txt', 'templates/_helpers.tpl'):
                (chart / name).write_bytes(tar.extractfile(prefix + name).read())
        (chart / 'Chart.yaml').write_text('apiVersion: v2\nname: notes-fixture\nversion: 0.1.0\n')
        defaults = yaml.safe_load((chart / 'values.yaml').read_text())

        def render(name, host='qa.example.test', port=0, kind='NodePort', enabled=True,
                   endpoint='', minio=True, sql_type='NodePort'):
            values = copy.deepcopy(defaults)
            values['access']['host'] = host
            values['sqlService']['type'] = sql_type
            values['dashboard']['enabled'] = enabled
            values['dashboard']['service'].update(type=kind, nodePort=port)
            values['objectStorage']['endpoint'] = endpoint
            values['minio']['enabled'] = minio
            (chart / 'values.yaml').write_text(yaml.safe_dump(values))
            output = run('helm', 'install', 'notes', str(chart), '-n', ns, '--dry-run=server')
            checks.append(name)
            notes = output.split('NOTES:', 1)[1]
            assert notes.isascii(), 'Expected English-only Helm NOTES'
            return notes

        run('kubectl', 'create', 'namespace', ns)
        try:
            first = render('first-install-auto-port')
            assert len(first.strip().splitlines()) == 10
            # Compare the complete install and saved NOTES, including the fallback
            # on a fresh release before Kubernetes assigns its Service ports.
            installed = run('helm', 'install', 'notes', str(chart), '-n', ns)
            assert installed.split('NOTES:', 1)[1].strip() == first.strip()
            saved = run('helm', 'get', 'notes', 'notes', '-n', ns)
            assert saved.split('NOTES:', 1)[-1].strip() == first.strip()
            run('helm', 'uninstall', 'notes', '-n', ns)
            checks.append('complete-install-and-saved-notes')

            def dashboard_command(notes):
                return next(line.strip() for line in notes.splitlines() if line.strip().startswith('DASHBOARD_PORT='))

            assert 'port-forward svc/oceanbase-dashboard-' not in first
            assert first.index(dashboard_command(first)) < first.index('get secret notes-dashboard-bootstrap')
            sql = next(line.strip() for line in first.splitlines() if line.strip().startswith('SQL_PORT='))
            assert f'kubectl -n {ns} get svc notes-sql' in sql
            for response in ('return 1', 'return 0'):
                failure = subprocess.run(['sh', '-c', 'kubectl() { ' + response + '; }\n'
                                          + 'mysql() { echo unexpected; }\n' + sql], capture_output=True)
                assert failure.returncode and not failure.stdout
            assert first.index(sql) < first.index('get secret notes-db-root') < first.index('OceanBase Operator Dashboard')
            internal = render('cluster-ip-sql', sql_type='ClusterIP')
            assert f"mysql -h notes-sql.{ns}.svc -P 2883 -u 'root@sys' -p" in internal
            assert "mysql -h 127.0.0.1 -P 2883 -u 'root@sys' -p" in internal
            assert 'get svc oceanbase-dashboard-notes' in first
            ipv6 = render('first-install-ipv6', host='2001:db8::1')
            assert "mysql -h '2001:db8::1' -P" in ipv6
            unsafe = render('host-shell-quoting', host="a'$(false)")
            assert "mysql -h 'a'\"'\"'$(false)'" in unsafe
            explicit = render('explicit-port-ipv6', host='2001:db8::1', port=31183)
            assert 'Dashboard URL: http://[2001:db8::1]:31183/#/login' in explicit
            service = {'apiVersion': 'v1', 'kind': 'Service',
                       'metadata': {'name': 'oceanbase-dashboard-notes', 'namespace': ns},
                       'spec': {'type': 'NodePort', 'ports': [{'port': 80, 'name': 'http'}]}}
            run('kubectl', 'apply', '-f', '-', input=json.dumps(service))
            actual = json.loads(run('kubectl', '-n', ns, 'get', 'svc',
                                    'oceanbase-dashboard-notes', '-o', 'json'))['spec']['ports'][0]['nodePort']
            for notes, host in [(first, 'qa.example.test'), (ipv6, '[2001:db8::1]')]:
                assert run('sh', '-ec', dashboard_command(notes)).strip() == f'http://{host}:{actual}/#/login'
            checks.append('dashboard-command-uses-allocated-port')
            sql_service = {'apiVersion': 'v1', 'kind': 'Service',
                           'metadata': {'name': 'notes-sql', 'namespace': ns},
                           'spec': {'type': 'NodePort', 'ports': [{'name': 'sql', 'port': 2883}]}}
            allocated = json.loads(run('kubectl', 'create', '-f', '-', '-o', 'json',
                                       input=json.dumps(sql_service)))['spec']['ports'][0]['nodePort']
            # Capture mysql's argv while letting kubectl read the real, newly allocated port.
            argv = run('sh', '-ec', 'mysql() { printf "%s\\n" "$@"; }\n' + sql).splitlines()
            assert argv == ['-h', 'qa.example.test', '-P', str(allocated), '-u', 'root@sys', '--protocol=TCP', '-p']
            checks.append('mysql-command-uses-allocated-port')
            existing = render('existing-service-auto-port')
            assert f"mysql -h 'qa.example.test' -P {allocated} -u 'root@sys' --protocol=TCP -p" in existing
            assert 'Dashboard URL: http://qa.example.test:%s/#/login' % actual in existing
            assert 'Dashboard URL: http://qa.example.test:31184/#/login' in render('explicit-overrides-existing', port=31184)
            assert 'port-forward svc/oceanbase-dashboard-notes' in render('cluster-ip', kind='ClusterIP')
            assert 'Dashboard LoadBalancer address' in render('load-balancer', kind='LoadBalancer')
            assert 'Dashboard is disabled' in render('dashboard-disabled', enabled=False)
            external = render('external-storage-no-minio', endpoint='https://s3.example.test', minio=False)
            assert external == existing
            separate = render('external-storage-with-minio', endpoint='https://s3.example.test')
            assert separate == existing
        finally:
            run('kubectl', 'delete', 'namespace', ns, '--wait=false')
    print(json.dumps({'status': 'PASS', 'checks': checks}, indent=2))


if __name__ == '__main__':
    main()

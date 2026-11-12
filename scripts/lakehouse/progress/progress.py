#!/usr/bin/env python3
"""Read-only Helm installation progress. Python 3.8+, kubectl and Helm 3."""
import argparse
from pathlib import Path
import datetime as dt
import json
import os
import re
import shlex
import shutil
import subprocess
import sys
import time

PHASES = ['Certificates and Operator', 'Object storage initialization', 'LogService', 'OceanBase', 'Dashboard', 'Final readiness checks']

def timestamp(value):
    try:
        return dt.datetime.fromisoformat(re.sub(r'(\.\d{6})\d+', r'\1', value).replace('Z', '+00:00')).timestamp()
    except (ValueError, AttributeError, TypeError):
        return None

def duration(seconds):
    seconds = max(0, int(seconds))
    return '%dm%02ds' % (seconds // 60, seconds % 60)

def name(obj):
    return obj.get('metadata', {}).get('name', '')

def condition(obj, kind):
    return any(c.get('type') == kind and str(c.get('status')).lower() == 'true'
               for c in obj.get('status', {}).get('conditions', []))

def cr_running(obj):
    return obj.get('status', {}).get('status') == 'running'

def deployment_ready(obj):
    spec, status = obj.get('spec', {}), obj.get('status', {})
    desired = spec.get('replicas', 1)
    return bool(obj) and desired > 0 and status.get('observedGeneration', 0) >= obj.get('metadata', {}).get('generation', 0) and status.get('updatedReplicas', 0) >= desired and status.get('availableReplicas', 0) >= desired

def owned_resources(items, release):
    # Follow UID ownership, including randomly named image-pull Jobs and their Pods.
    selected = [x for x in items if
                x.get('metadata', {}).get('labels', {}).get('app.kubernetes.io/instance') == release
                or x.get('metadata', {}).get('labels', {}).get('ref-obcluster') == release + '-ob'
                or x.get('metadata', {}).get('labels', {}).get('ref-oblogservicecluster') == release + '-ls'
                or name(x) in [release + '-ob', release + '-ls', release + '-prerequisites', release + '-readiness', 'oceanbase-dashboard-' + release]
                or name(x).startswith(release + '-storage-init-')]
    seen = {x.get('metadata', {}).get('uid') for x in selected} - {None}
    while True:
        extra = [x for x in items if x not in selected and any(
            owner.get('uid') in seen for owner in x.get('metadata', {}).get('ownerReferences', []))]
        if not extra:
            break
        selected += extra
        seen.update(x.get('metadata', {}).get('uid') for x in extra if x.get('metadata', {}).get('uid'))
    return selected

def analyze(release, helm, items, crs):
    resources = owned_resources(items + crs, release)
    jobs = {name(x): x for x in resources if x.get('kind') == 'Job'}
    pods = [x for x in resources if x.get('kind') == 'Pod']
    by_name = {name(x): x for x in resources}
    info = helm.get('info', {})
    status = info.get('status', 'not-found')
    revision = helm.get('version', 1)
    pre = jobs.get(release + '-prerequisites', {})
    storage = jobs.get('%s-storage-init-%s' % (release, revision), {})
    ls, ob = by_name.get(release + '-ls', {}), by_name.get(release + '-ob', {})
    dashboard = by_name.get('oceanbase-dashboard-' + release, {})
    pre_active = bool(pre) and not condition(pre, 'Complete')
    passed = [condition(pre, 'Complete') or (bool(ls or ob or storage) and not pre_active),
              condition(storage, 'Complete') or cr_running(ls), cr_running(ls),
              cr_running(ob), deployment_ready(dashboard), status == 'deployed']
    if pre_active:
        stage = 0
    else:
        stage = next((i for i, ok in enumerate(passed) if not ok), 5)
        # External storage/disabled Dashboard need no corresponding resource.
        # Do not claim absent resources succeeded; Helm's final result is authoritative.
        if not storage and (ls or ob):
            stage = 2 if not cr_running(ls) else (3 if not cr_running(ob) else 5)
        if stage == 4 and not dashboard:
            stage = 5
    uid_set = {x.get('metadata', {}).get('uid') for x in resources} - {None}
    events = [x for x in items if x.get('kind') == 'Event' and
              x.get('involvedObject', {}).get('uid') in uid_set]
    def event_time(e):
        return timestamp(e.get('series', {}).get('lastObservedTime') or e.get('lastTimestamp') or e.get('eventTime') or e.get('metadata', {}).get('creationTimestamp')) or 0
    events.sort(key=event_time)
    latest = events[-1] if events else None
    issues = []
    for cr in crs:
        if cr.get('status', {}).get('status') == 'failed':
            issues.append('Cluster initialization failed: ' + name(cr))
    for job in jobs.values():
        if condition(job, 'Failed'):
            issues.append('Job failed: ' + name(job))
    for pod in pods:
        if pod.get('status', {}).get('phase') == 'Failed':
            issues.append('Pod failed: ' + name(pod))
        for c in pod.get('status', {}).get('initContainerStatuses', []) + pod.get('status', {}).get('containerStatuses', []):
            waiting = c.get('state', {}).get('waiting', {})
            if waiting.get('reason') in ['ImagePullBackOff', 'ErrImagePull', 'CrashLoopBackOff', 'CreateContainerConfigError', 'CreateContainerError', 'InvalidImageName']:
                issues.append('%s: %s %s' % (name(pod), waiting['reason'], waiting.get('message', '')))
        for c in pod.get('status', {}).get('conditions', []):
            if c.get('type') == 'PodScheduled' and c.get('status') == 'False':
                issues.append('%s: %s' % (name(pod), c.get('message', 'Waiting for scheduling')))
    detail = 'Waiting for resources or status updates'
    for pod in pods:
        statuses = pod.get('status', {}).get('containerStatuses', [])
        if name(pod).startswith('image-pull-ready-') and pod.get('status', {}).get('phase') not in ['Succeeded', 'Failed']:
            detail = 'Preparing the Observer image (pulling, unpacking or checking)'
            break
        if any(c.get('state', {}).get('waiting', {}).get('reason') == 'ContainerCreating' for c in statuses):
            detail = 'Creating container: ' + name(pod)
    if stage == 0:
        detail = 'Preparing certificates and Operator; waiting for the prerequisites Job' if pre else 'Waiting for Helm to register CRDs and create the prerequisites Job'
    if stage == 3 and ob and detail == 'Waiting for resources or status updates':
        detail = 'Initializing the database; cluster status: ' + str(ob.get('status', {}).get('status', 'unknown'))
    if stage == 5:
        detail = 'Waiting for Helm final readiness checks'
    counts = {}
    for label, prefix in [('LogService', release + '-ls-'), ('OceanBase', release + '-ob-')]:
        service_pods = [p for p in pods if name(p).startswith(prefix) and not any(o.get('kind') == 'Job' for o in p.get('metadata', {}).get('ownerReferences', []))]
        if service_pods:
            counts[label] = '%s/%s Pods ready' % (sum(condition(p, 'Ready') for p in service_pods), len(service_pods))
    return dict(status=status, stage=stage, passed=passed, detail=detail, issues=issues,
                latest=latest, latest_time=event_time(latest) if latest else 0, counts=counts,
                started=timestamp(info.get('last_deployed')), description=info.get('description', ''))

def default_kubeconfig():
    candidates = []
    configured = os.environ.get('KUBECONFIG')
    if configured:
        candidates.extend(Path(item).expanduser() for item in configured.split(os.pathsep) if item)
    else:
        candidates.append(Path.home() / '.kube' / 'config')
    candidates.append(Path('/etc/rancher/k3s/k3s.yaml'))
    for candidate in candidates:
        try:
            if candidate.is_file() and os.access(candidate, os.R_OK):
                return str(candidate)
        except OSError:
            pass
    return None

def execute(argv):
    try:
        p = subprocess.run(argv, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, timeout=12)
    except subprocess.TimeoutExpired:
        raise RuntimeError('Query timed out: ' + shlex.join(argv[:3]))
    if p.returncode:
        raise RuntimeError(p.stderr.strip() or 'Query failed')
    try:
        return json.loads(p.stdout)
    except ValueError:
        raise RuntimeError('Command returned invalid JSON: ' + shlex.join(argv[:3]))

def snapshot(args):
    flags = []
    hflags = []
    kubeconfig = args.kubeconfig or default_kubeconfig()
    if args.context:
        flags += ['--context', args.context]
        hflags += ['--kube-context', args.context]
    if kubeconfig:
        flags += ['--kubeconfig', kubeconfig]
        hflags += ['--kubeconfig', kubeconfig]
    k = ['kubectl'] + flags + ['--request-timeout=8s', '-n', args.namespace]
    try:
        helm = execute(['helm'] + hflags + ['status', args.release, '-n', args.namespace, '-o', 'json'])
    except RuntimeError as e:
        if 'release: not found' not in str(e).lower():
            raise
        helm = {}
    items = execute(k + ['get', 'pods,jobs,deployments,events', '-o', 'json']).get('items', [])
    crs = []
    for resource, suffix in [('oblogserviceclusters.oceanbase.oceanbase.com', '-ls'), ('obclusters.oceanbase.oceanbase.com', '-ob')]:
        try:
            obj = execute(k + ['get', resource, args.release + suffix, '-o', 'json'])
            crs.append(obj)
        except RuntimeError as e:
            message = str(e).lower()
            if not any(s in message for s in ['(notfound)', "doesn't have a resource type", 'could not find the requested resource']):
                raise
    return analyze(args.release, helm, items, crs)

def render(args, state, stage_start, watching_start):
    now = time.time()
    lines = ['Lakehouse installation progress  namespace=%s  release=%s' % (args.namespace, args.release),
             'Helm status: ' + state['status'], '']
    for i, ok in enumerate(state['passed'][:-1]):
        if ok:
            lines.append('✓ ' + PHASES[i] + ' ready' + (': ' + state['counts'][PHASES[i]] if PHASES[i] in state['counts'] else ''))
    if state['status'] == 'deployed':
        lines += ['', '✓ Helm installation completed (deployed)']
    elif state['status'] == 'failed':
        lines += ['', '✗ Helm installation failed: ' + state['description']]
    else:
        lines += ['', '[%d/6] %s' % (state['stage'] + 1, PHASES[state['stage']]), '  Current: ' + state['detail']]
    total = ('Installation elapsed: ' + duration(now - state['started'])) if state['started'] else ('Observation elapsed: ' + duration(now - watching_start))
    if state['status'] not in ['deployed', 'failed']:
        lines += ['  Phase observed: %s  %s' % (duration(now - stage_start), total)]
    if state['latest']:
        event = state['latest']
        lines.append('  Latest event (%s ago): %s %s' % (duration(now - state['latest_time']), event.get('reason', ''), event.get('message', '')[:700]))
    for issue in state['issues'][:5]:
        lines.append('  ⚠ ' + issue[:700])
    if state['status'].startswith('pending') and not state['issues']:
        lines.append('  Waiting; no explicit failure detected in this snapshot.')
    if state['status'] in ['deployed', 'failed']:
        lines += ['', 'Viewer finished; no cluster resources were modified.']
    else:
        lines += ['', 'Read-only viewer; refresh every %s seconds. Ctrl+C exits the viewer without stopping Helm.' % args.interval]
    return '\n'.join(lines)

def main_with_args(args):
    watching_start = stage_start = time.time()
    previous = None
    while True:
        cycle_started = time.monotonic()
        try:
            state = snapshot(args)
            key = (state['started'], state['stage'])
            if key != previous:
                stage_start = time.time()
                previous = key
            if sys.stdout.isatty() and not args.no_clear:
                print('\033[2J\033[H', end='')
            print(render(args, state, stage_start, watching_start), flush=True)
            if state['status'] in ['deployed', 'failed', 'uninstalled', 'superseded']:
                return 0 if state['status'] == 'deployed' else 1
            if args.once:
                return 0
        except RuntimeError as e:
            message = str(e)
            if 'permission denied' in message.lower() and '/etc/rancher/k3s/k3s.yaml' in message:
                message += (
                    '\nK3s kubeconfig is not readable by the current user. Ask an administrator to run:\n'
                    '  sudo install -d -m 700 -o "$USER" -g "$(id -gn "$USER")" "$HOME/.kube"\n'
                    '  sudo install -m 600 -o "$USER" -g "$(id -gn "$USER")" /etc/rancher/k3s/k3s.yaml "$HOME/.kube/config"\n'
                    'Then rerun this viewer in the same terminal.'
                )
            print('Unable to read status (this does not mean installation failed): ' + message, file=sys.stderr, flush=True)
            if args.once:
                return 2
        if args.timeout and time.time() - watching_start >= args.timeout:
            print('Viewer timed out; Helm installation was not interrupted.', file=sys.stderr)
            return 2
        time.sleep(max(0, args.interval - (time.monotonic() - cycle_started)))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('-n', '--namespace', default='ob-demo')
    parser.add_argument('--release', default='ob-demo')
    parser.add_argument('--interval', type=int, default=3)
    parser.add_argument('--once', action='store_true', help='Print one snapshot and exit')
    parser.add_argument('--no-clear', action='store_true', help='Append snapshots instead of refreshing the terminal')
    parser.add_argument('--timeout', type=int, default=0, help='Viewer timeout in seconds; 0 means unlimited')
    parser.add_argument('--context')
    parser.add_argument('--kubeconfig')
    args = parser.parse_args()
    if args.interval < 1 or args.timeout < 0:
        parser.error('interval must be >=1; timeout must be >=0')
    for command in ['kubectl', 'helm']:
        if not shutil.which(command):
            parser.error('Required command not found: ' + command)
    return main_with_args(args)


if __name__ == '__main__':
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        print('\nExited the viewer; Helm installation was not stopped.')
        sys.exit(130)

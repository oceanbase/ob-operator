"""Wait for the deployments and admission endpoints from one bootstrap stage."""
import json
import os
from pathlib import Path
import ssl
import sys
import time
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


SA = Path('/var/run/secrets/kubernetes.io/serviceaccount')
TLS = ssl.create_default_context(cafile=str(SA / 'ca.crt'))
HOST = os.environ['KUBERNETES_SERVICE_HOST']
if ':' in HOST:
    HOST = '[' + HOST + ']'
API = 'https://%s:%s' % (HOST, os.environ['KUBERNETES_SERVICE_PORT'])


def get(path):
    token = (SA / 'token').read_text().strip()
    request = Request(API + path, headers={'Authorization': 'Bearer ' + token})
    with urlopen(request, context=TLS, timeout=5) as response:
        return json.load(response)


def condition(obj, name):
    return any(item['type'] == name and item['status'] == 'True'
               for item in obj.get('status', {}).get('conditions', []))


def available(deployment):
    return (deployment.get('status', {}).get('observedGeneration', 0) >= deployment['metadata']['generation']
            and condition(deployment, 'Available'))


def ready(stage):
    namespace = stage['namespace']
    for name in stage['deployments']:
        if not available(get('/apis/apps/v1/namespaces/%s/deployments/%s' % (namespace, name))):
            return 'Waiting for Deployment ' + name
    endpoints = get('/api/v1/namespaces/%s/endpoints/%s' % (namespace, stage['service']))
    if not any(subset.get('addresses') for subset in endpoints.get('subsets', [])):
        return 'Waiting for webhook endpoints'
    if stage.get('certificate') and not condition(
            get('/apis/cert-manager.io/v1/namespaces/%s/certificates/%s' % (namespace, stage['certificate'])), 'Ready'):
        return 'Waiting for serving Certificate'
    for kind in ('mutating', 'validating'):
        path = '/apis/admissionregistration.k8s.io/v1/%swebhookconfigurations/%s' % (kind, stage[kind])
        webhooks = get(path).get('webhooks', [])
        if not webhooks or not all(item['clientConfig'].get('caBundle') for item in webhooks):
            return 'Waiting for CA injection into ' + stage[kind]
    return None


stage = json.loads(sys.argv[1])
previous = None
while True:
    try:
        message = ready(stage)
    except HTTPError as exc:
        if exc.code not in (404, 429, 500, 502, 503, 504):
            raise
        message = '%s returned HTTP %d' % (exc.filename, exc.code)
    except (URLError, TimeoutError) as exc:
        message = str(exc)
    if message is None:
        print(stage['name'] + ' prerequisites are ready')
        break
    if message != previous:
        print(message)
        previous = message
    time.sleep(3)

"""Delete this Helm release's OBCluster before its LogService."""
import json
import os
from pathlib import Path
import ssl
import sys
import time
from urllib.error import HTTPError
from urllib.request import Request, urlopen


namespace, release = sys.argv[1:]
account = Path('/var/run/secrets/kubernetes.io/serviceaccount')
tls = ssl.create_default_context(cafile=str(account / 'ca.crt'))
host = os.environ['KUBERNETES_SERVICE_HOST']
if ':' in host:
    host = '[' + host + ']'
api = 'https://%s:%s/apis/oceanbase.oceanbase.com/v1alpha1/namespaces/%s' % (
    host, os.environ['KUBERNETES_SERVICE_PORT'], namespace)


def request(path, options=None):
    headers = {'Authorization': 'Bearer ' + (account / 'token').read_text().strip()}
    data = None
    if options is not None:
        headers['Content-Type'] = 'application/json'
        data = json.dumps(options).encode()
    req = Request(api + '/' + path, data=data, headers=headers,
                  method='GET' if options is None else 'DELETE')
    try:
        with urlopen(req, context=tls, timeout=10) as response:
            return json.load(response)
    except HTTPError as error:
        if error.code == 404:
            return None
        raise RuntimeError('%s %s: %s' % (req.method, path, error.read().decode())) from error


def delete_owned(path):
    obj = request(path)
    if obj is None:
        return
    meta = obj['metadata']
    owner = meta.get('annotations', {})
    if (owner.get('meta.helm.sh/release-name') != release or
            owner.get('meta.helm.sh/release-namespace') != namespace):
        print('Skipping %s owned outside this release' % path)
        return
    if not meta.get('deletionTimestamp'):
        print('Deleting ' + path)
        request(path, {'preconditions': {'uid': meta['uid'], 'resourceVersion': meta['resourceVersion']},
                       'propagationPolicy': 'Background'})
    while True:
        remaining = request(path)
        if remaining is None or remaining['metadata']['uid'] != meta['uid']:
            return
        time.sleep(2)


delete_owned('obclusters/' + release + '-ob')
delete_owned('oblogserviceclusters/' + release + '-ls')

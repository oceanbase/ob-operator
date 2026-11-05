import json,os,pathlib,subprocess,tempfile,unittest
HERE=pathlib.Path(__file__).resolve().parent
class TimeoutTests(unittest.TestCase):
 def test_operator_wait_uses_configured_budget(self):
  for seconds in [None,'1200']:
   with self.subTest(seconds=seconds),tempfile.TemporaryDirectory() as tmp:
    root=pathlib.Path(tmp);log=root/'calls';stub=root/'k3s'
    stub.write_text('''#!/usr/bin/env python3
import os,sys,json
with open(os.environ['CALL_LOG'],'a') as f:f.write(json.dumps(sys.argv[1:])+'\\n')
a=' '.join(sys.argv)
if 'get deployments' in a:print('oceanbase-system allowed-image')
elif 'get endpoints' in a:print('10.42.0.3')
elif 'get validatingwebhookconfiguration' in a or 'get mutatingwebhookconfiguration' in a:print('CA')
''');stub.chmod(0o755)
    (root/'operator-compatible-images.txt').write_text('allowed-image\n')
    script=(HERE/'prerequisites/prepare.sh').read_text().replace('/bin/k3s',str(stub)).replace('/payload',str(root)).replace('/tmp/prerequisites-kubeconfig',str(root/'kubeconfig'))
    path=root/'prepare.sh';path.write_text(script)
    env=dict(os.environ,CALL_LOG=str(log),BOOTSTRAP_BINDING='test-binding')
    if seconds:env['WAIT_TIMEOUT_SECONDS']=seconds
    else:env.pop('WAIT_TIMEOUT_SECONDS',None)
    p=subprocess.run(['/bin/sh',str(path)],env=env,capture_output=True,text=True)
    self.assertEqual(p.returncode,0,p.stderr)
    calls=[json.loads(x) for x in log.read_text().splitlines()]
    rollout=next(x for x in calls if 'rollout' in x)
    self.assertIn('--timeout='+str(seconds or 1800)+'s',rollout)
if __name__=='__main__':unittest.main()

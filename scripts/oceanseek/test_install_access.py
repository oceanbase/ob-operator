"""Run the post-install address wrapper with a controlled Helm/Kubectl boundary."""
import json, os, pathlib, subprocess, tempfile, unittest
SCRIPT=pathlib.Path(__file__).with_name('install-helm-and-show-access.sh')
class AccessTest(unittest.TestCase):
 def test_context_host_and_failure(self):
  with tempfile.TemporaryDirectory() as tmp:
   p=pathlib.Path(tmp)
   stub='''#!/usr/bin/env python3
import sys,os,json
name=os.path.basename(sys.argv[0]);args=sys.argv[1:]
with open(os.environ['CALLS'],'a') as f:f.write(json.dumps([name,args])+'\\n')
if name=='helm':
 if args[:2]==['upgrade','--install']:sys.exit(int(os.environ.get('FAIL','0')))
 print('{"access":{"host":"2001:db8::1"}}')
else:
 arg=' '.join(args)
 if 'spec.type' in arg:print('NodePort')
 elif 'nodePort' in arg:print('31080' if 'dashboard' in arg else '31081')
 elif 'ports[0].port' in arg:print('8088')
'''
   for n in ['helm','kubectl']:(p/n).write_text(stub);(p/n).chmod(0o755)
   env=dict(os.environ,PATH=str(p)+':'+os.environ['PATH'],CALLS=str(p/'calls'));env.pop('OCEANSEEK_ACCESS_HOST',None)
   cmd=['bash',str(SCRIPT),'demo','chart.tgz','-n','demo','--kubeconfig','/test/private config','--kube-context=target']
   r=subprocess.run(cmd,env=env,text=True,capture_output=True);self.assertEqual(r.returncode,0,r.stderr)
   self.assertIn('NodePort endpoint: [2001:db8::1]:31080',r.stdout)
   self.assertNotIn('http://[2001:db8::1]',r.stdout)
   self.assertIn('port-forward svc/oceanbase-dashboard-demo 8080:8088',r.stdout)
   self.assertIn('Open http://127.0.0.1:8080',r.stdout)
   self.assertIn("mysql -h '2001:db8::1' -P 31081",r.stdout)
   calls=[json.loads(l) for l in (p/'calls').read_text().splitlines()]
   for name,args in calls:
    self.assertIn('/test/private config',args)
    if name=='kubectl':self.assertIn('--context=target',args)
   (p/'calls').write_text('');r=subprocess.run(cmd,env=dict(env,FAIL='7'),text=True,capture_output=True)
   self.assertIn('storage initialization',r.stderr);self.assertIn('app.kubernetes.io/component=storage-init',r.stderr);self.assertIn('/test/private\\ config',r.stderr);self.assertIn('--context=target',r.stderr);self.assertEqual(r.returncode,7);self.assertNotIn('Dashboard URL:',r.stdout);self.assertEqual(len((p/'calls').read_text().splitlines()),1)
if __name__=='__main__':unittest.main()

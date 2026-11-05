#!/usr/bin/env python3
"""Build the single-command Helm test delivery from a verified online bundle."""
import argparse,hashlib,importlib.util,json,shutil,subprocess,tarfile,tempfile
from pathlib import Path
import yaml
HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('native',HERE/'build-native-packages.py')
native=importlib.util.module_from_spec(spec);spec.loader.exec_module(native)
def run(args):return subprocess.check_output(list(map(str,args)),text=True)
def build(args):
 source=HERE.parent.parent;baseline=args.baseline.resolve();output=args.output.resolve()
 output.mkdir(parents=True,exist_ok=False)
 bundle=output/('oceanbase-lakehouse-native-helm-'+args.version)
 shutil.copytree(baseline,bundle)
 for p in bundle.glob('oceanbase-lakehouse-*.tgz'):p.unlink()
 lock=json.loads((baseline/'image-lock.json').read_text());extra=json.loads(args.extra_images.read_text())
 for item in extra:
  if not item.get('registry_pull_verified'):raise ValueError('Unverified extra image '+item['role'])
  item['channels']=['helm'];item['publication']='verified-registry-pull'
 lock['images']+=extra;lock['delivery_version']=args.version
 lock['prerequisites']=['Kubernetes >=1.27, Helm 3, cluster-administrator install access, working StorageClass (local-path by default)']
 refs={i['role']:i['reference'] for i in lock['images']}
 with tempfile.TemporaryDirectory() as tmp:
  tmp=Path(tmp)
  original,=baseline.glob('oceanbase-lakehouse-*.tgz')
  with tarfile.open(original) as archive:
   for member in archive:
    if member.name.startswith('/') or '..' in Path(member.name).parts or member.issym() or member.islnk() or not (member.isfile() or member.isdir()):raise ValueError('Unsafe chart archive')
   archive.extractall(tmp)
  chart=tmp/'oceanbase-lakehouse'
  for name in ['_helpers.tpl','prerequisites.yaml','NOTES.txt']:
   shutil.copyfile(source/'charts/oceanbase-lakehouse/templates'/name,chart/'templates'/name)
  shutil.copyfile(source/'charts/oceanbase-lakehouse/values.schema.json',chart/'values.schema.json')
  shutil.copytree(HERE/'prerequisites/crds',chart/'crds')
  payload=chart/'files/prerequisites';payload.mkdir(parents=True)
  shutil.copytree(HERE/'prerequisites/manifests',payload,dirs_exist_ok=True)
  shutil.copyfile(HERE/'prerequisites/prepare.sh',payload/'prepare.sh')
  text=(payload/'cert-manager.yaml').read_text()
  for component in ['controller','cainjector','webhook']:
   text=text.replace('quay.io/jetstack/cert-manager-'+component+':v1.12.4',refs['cert-manager-'+component])
  (payload/'cert-manager.yaml').write_text(text)
  allowed=next(i for i in lock['images'] if i['role']=='operator')
  (payload/'operator-compatible-images.txt').write_text('\n'.join(dict.fromkeys([allowed['reference'],allowed.get('upstream_reference',''),allowed.get('source_reference','')]))+'\n')
  values=yaml.safe_load((chart/'values.yaml').read_text());values['prerequisites']={'enabled':True,'image':refs['prerequisites'],'timeoutSeconds':1800}
  (chart/'values.yaml').write_text(yaml.safe_dump(values,sort_keys=False))
  meta=yaml.safe_load((chart/'Chart.yaml').read_text());meta['version']='0.5.0';(chart/'Chart.yaml').write_text(yaml.safe_dump(meta,sort_keys=False))
  run(['helm','lint','--strict',chart]);rendered=run(['helm','template','ob-demo',chart,'-n','ob-demo','--include-crds'])
  (output/'rendered-validation.yaml').write_text(rendered)
  run(['helm','package',chart,'--destination',bundle])
 shutil.copyfile(HERE/'package-offline.py',bundle/'tools/package-offline.py')
 for name in ['progress.sh','progress.py']:
  shutil.copy2(HERE/'progress'/name,bundle/'tools'/name)
 (bundle/'image-lock.json').write_text(json.dumps(lock,indent=2)+'\n')
 (bundle/'release-images.lock.json').write_text(json.dumps(lock,indent=2)+'\n')
 image_values=yaml.safe_load((bundle/'image-values.yaml').read_text()) or {}
 image_values['prerequisites']={'enabled':True,'image':refs['prerequisites']}
 (bundle/'image-values.yaml').write_text(yaml.safe_dump(image_values,sort_keys=False))
 shutil.copyfile(baseline/'USAGE.txt',bundle/'ADVANCED-USAGE.txt')
 instructions='''# 单条 Helm 命令安装湖库与 Dashboard

前提：已安装 K3s/Kubernetes >=1.27、Helm 3，具有集群管理员权限，节点能访问镜像仓库。
K3s 默认 local-path 存储可直接使用；其他集群追加 --set storageClass=实际存储类。
不再要求用户手动安装 Operator 或 cert-manager。

从解压目录执行：

```bash
helm upgrade --install ob-demo ./oceanbase-lakehouse-0.5.0.tgz -n ob-demo --create-namespace --wait --wait-for-jobs --timeout 35m
```

Chart 先注册 CRD，再通过安装前 Job 准备 cert-manager 和匹配的 Operator，并等待证书和
webhook 就绪。之后安装存储、LogService、OceanBase 和 Dashboard。不要加 --no-hooks。
预置的 cert-manager 是 v1.12.4（与当前测试环境一致），不是对任意现有版本的升级请求。
安装期间建议另开终端查看统一进度（需要 Python 3.8+、kubectl、Helm 3）：

```bash
./tools/progress.sh -n ob-demo --release ob-demo
```

默认每 3 秒刷新，以英文显示阶段、总耗时、阶段观察耗时、资源等待原因和最近事件。
Ctrl+C 只退出查看；不会停止安装。安装完成或失败后自动退出。
可用 --once 打印单次快照，--no-clear 保留滚动记录，--interval 5 调整刷新间隔。
支持 --context 和 --kubeconfig；两个终端应使用同一集群。

也可以单独查看任务日志：

```bash
kubectl -n ob-demo logs -f job/ob-demo-prerequisites
kubectl -n ob-demo logs -f job/ob-demo-readiness
```

成功的 Job 会被删除，因此完成后查询可能显示 NotFound。失败时保留日志，同一命令可重试。
安装前任务执行时临时具有安装集群组件的权限，退出时撤销自己的 ClusterRoleBinding。
如果镜像拉取/调度失败，任务没有启动，需在排障后重试；也可主动删除
`ob-demo-ob-demo-prerequisites` ClusterRoleBinding 撤销尚未使用的权限。

已有匹配 Operator 时复用；不同镜像版本、多套 Operator、非标准 namespace 的 cert-manager
或其他 Helm release 的遗留 Operator 资源会明确报错，不自动覆盖。
不要同时启动两个首次安装任务。已有共享基础组件由管理员维护，本包不自动升级它们。

CRD、cert-manager 和 Operator 是共享集群基础组件，卸载 ob-demo 时保留；数据库的数据保留
规则同原包。不要通过删除 namespace 卸载。切换到已由管理员管理的环境，可设置
--set prerequisites.enabled=false，但此时管理员必须提前保证 Operator 和 webhook 已就绪。

Dashboard 默认映射 NodePort 31183，SQL 为 30881，Helm 输出访问 IP 与取密码命令。
初始化账户是 Dashboard 的 admin，与数据库 root@sys 分开。

离线工具见 OFFLINE.md；镜像清单包括前置任务和 cert-manager 的三个镜像。
'''
 for name in ['README.md','QUICKSTART.md','USAGE.txt','IMAGE-DELIVERY.txt']:(bundle/name).write_text(instructions)
 offline=(bundle/'OFFLINE.md').read_text()
 offline=offline.replace('或 cert-manager','').replace('、StorageClass 或 cert-manager','、StorageClass')
 offline+='\n本单命令 Helm 包已内置证书组件和 Operator 的自动准备流程；无需单独安装。\n离线镜像清单包含其镜像及前置 Job 镜像，须导入每个可能运行这些组件的节点。\n'
 (bundle/'OFFLINE.md').write_text(offline)
 validation={'version':args.version,'chart_version':'0.5.0','helm_lint':True,'helm_template':True,'live_business_installation_tested':False,'registry_images':len(lock['images'])}
 if args.validation:
  validation.update(json.loads(args.validation.read_text()))
 (bundle/'build-validation.json').write_text(json.dumps(validation,indent=2)+'\n')
 checks=[native.sha256(p)+'  '+p.relative_to(bundle).as_posix() for p in sorted(bundle.rglob('*')) if p.is_file() and p.name!='SHA256SUMS']
 (bundle/'SHA256SUMS').write_text('\n'.join(checks)+'\n')
 target=output/(bundle.name+'.tar.gz');native.tar_tree(bundle,target)
 (output/'SHA256SUMS').write_text(native.sha256(target)+'  '+target.name+'\n')
 print(json.dumps({'bundle':str(bundle),'archive':str(target),'bytes':target.stat().st_size}))
if __name__=='__main__':
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--baseline',type=Path,required=True);p.add_argument('--extra-images',type=Path,required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--version',required=True);p.add_argument('--validation',type=Path);build(p.parse_args())

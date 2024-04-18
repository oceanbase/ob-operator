"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[3134],{8578:(e,s,n)=>{n.r(s),n.d(s,{assets:()=>t,contentTitle:()=>o,default:()=>p,frontMatter:()=>c,metadata:()=>i,toc:()=>d});var r=n(4848),a=n(8453),l=n(8774);const c={title:"\u9879\u76ee\u4ecb\u7ecd"},o="ob-operator",i={type:"mdx",permalink:"/ob-operator/zh-Hans/",source:"@site/i18n/zh-Hans/docusaurus-plugin-content-pages/index.mdx",title:"\u9879\u76ee\u4ecb\u7ecd",description:"ob-operator \u662f\u6ee1\u8db3 Kubernetes Operator \u6269\u5c55\u8303\u5f0f\u7684\u81ea\u52a8\u5316\u5de5\u5177\uff0c\u53ef\u4ee5\u6781\u5927\u7b80\u5316\u5728 Kubernetes \u4e0a\u90e8\u7f72\u548c\u7ba1\u7406 OceanBase \u96c6\u7fa4\u53ca\u76f8\u5173\u8d44\u6e90\u7684\u8fc7\u7a0b\u3002",frontMatter:{title:"\u9879\u76ee\u4ecb\u7ecd"},unlisted:!1},t={},d=[{value:"\u5feb\u901f\u4e0a\u624b",id:"\u5feb\u901f\u4e0a\u624b",level:2},{value:"\u524d\u63d0\u6761\u4ef6",id:"\u524d\u63d0\u6761\u4ef6",level:3},{value:"\u90e8\u7f72 ob-operator",id:"\u90e8\u7f72-ob-operator",level:3},{value:"\u4f7f\u7528 YAML \u914d\u7f6e\u6587\u4ef6\xb7",id:"\u4f7f\u7528-yaml-\u914d\u7f6e\u6587\u4ef6",level:4},{value:"\u4f7f\u7528 Helm Chart",id:"\u4f7f\u7528-helm-chart",level:4},{value:"\u4f7f\u7528 terraform",id:"\u4f7f\u7528-terraform",level:4},{value:"\u9a8c\u8bc1\u90e8\u7f72\u7ed3\u679c",id:"\u9a8c\u8bc1\u90e8\u7f72\u7ed3\u679c",level:4},{value:"\u90e8\u7f72 OceanBase \u96c6\u7fa4",id:"\u90e8\u7f72-oceanbase-\u96c6\u7fa4",level:3},{value:"\u8fde\u63a5\u96c6\u7fa4",id:"\u8fde\u63a5\u96c6\u7fa4",level:3},{value:"OceanBase Dashboard",id:"oceanbase-dashboard",level:3},{value:"\u9879\u76ee\u67b6\u6784",id:"\u9879\u76ee\u67b6\u6784",level:2},{value:"\u7279\u6027",id:"\u7279\u6027",level:2},{value:"\u652f\u6301\u7684 OceanBase \u7248\u672c",id:"\u652f\u6301\u7684-oceanbase-\u7248\u672c",level:2},{value:"\u73af\u5883\u4f9d\u8d56",id:"\u73af\u5883\u4f9d\u8d56",level:2},{value:"\u6587\u6863",id:"\u6587\u6863",level:2},{value:"\u83b7\u53d6\u5e2e\u52a9",id:"\u83b7\u53d6\u5e2e\u52a9",level:2},{value:"\u53c2\u4e0e\u5f00\u53d1",id:"\u53c2\u4e0e\u5f00\u53d1",level:2},{value:"\u8bb8\u53ef\u8bc1",id:"\u8bb8\u53ef\u8bc1",level:2}];function h(e){const s={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",h4:"h4",img:"img",input:"input",li:"li",ol:"ol",p:"p",pre:"pre",ul:"ul",...(0,a.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(s.h1,{id:"ob-operator",children:"ob-operator"}),"\n",(0,r.jsxs)(s.p,{children:[(0,r.jsx)(s.code,{children:"ob-operator"})," \u662f\u6ee1\u8db3 Kubernetes Operator \u6269\u5c55\u8303\u5f0f\u7684\u81ea\u52a8\u5316\u5de5\u5177\uff0c\u53ef\u4ee5\u6781\u5927\u7b80\u5316\u5728 Kubernetes \u4e0a\u90e8\u7f72\u548c\u7ba1\u7406 OceanBase \u96c6\u7fa4\u53ca\u76f8\u5173\u8d44\u6e90\u7684\u8fc7\u7a0b\u3002"]}),"\n",(0,r.jsx)(s.h2,{id:"\u5feb\u901f\u4e0a\u624b",children:"\u5feb\u901f\u4e0a\u624b"}),"\n",(0,r.jsx)(s.p,{children:"\u8fd9\u90e8\u5206\u4ee5\u4e00\u4e2a\u7b80\u5355\u793a\u4f8b\u8bf4\u660e\u5982\u4f55\u4f7f\u7528 ob-operator \u5feb\u901f\u90e8\u7f72 OceanBase \u96c6\u7fa4\u3002"}),"\n",(0,r.jsx)(s.h3,{id:"\u524d\u63d0\u6761\u4ef6",children:"\u524d\u63d0\u6761\u4ef6"}),"\n",(0,r.jsx)(s.p,{children:"\u5f00\u59cb\u4e4b\u524d\u8bf7\u51c6\u5907\u4e00\u5957\u53ef\u7528\u7684 Kubernetes \u96c6\u7fa4\uff0c\u5e76\u4e14\u81f3\u5c11\u53ef\u4ee5\u5206\u914d 2C, 10G \u5185\u5b58\u4ee5\u53ca 100G \u5b58\u50a8\u7a7a\u95f4\u3002"}),"\n",(0,r.jsxs)(s.p,{children:["ob-operator \u4f9d\u8d56 ",(0,r.jsx)(s.a,{href:"https://cert-manager.io/docs/",children:"cert-manager"}),", cert-manager \u7684\u5b89\u88c5\u53ef\u4ee5\u53c2\u8003\u5bf9\u5e94\u7684",(0,r.jsx)(s.a,{href:"https://cert-manager.io/docs/installation/",children:"\u5b89\u88c5\u6587\u6863"}),"\uff0c\u5982\u679c\u60a8\u65e0\u6cd5\u8bbf\u95ee\u5b98\u65b9\u5236\u54c1\u6258\u7ba1\u5728 ",(0,r.jsx)(s.code,{children:"quay.io"})," \u955c\u50cf\u7ad9\u7684\u955c\u50cf\uff0c\u53ef\u901a\u8fc7\u4e0b\u9762\u7684\u6307\u4ee4\u5b89\u88c5\u6211\u4eec\u8f6c\u6258\u5728 ",(0,r.jsx)(s.code,{children:"docker.io"})," \u4e2d\u7684\u5236\u54c1\uff1a"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.2.0_release/deploy/cert-manager.yaml\n"})}),"\n",(0,r.jsxs)(s.p,{children:["\u672c\u4f8b\u5b50\u4e2d\u7684 OceanBase \u96c6\u7fa4\u5b58\u50a8\u4f9d\u8d56 ",(0,r.jsx)(s.a,{href:"https://github.com/rancher/local-path-provisioner",children:"local-path-provisioner"})," \u63d0\u4f9b, \u9700\u8981\u63d0\u524d\u8fdb\u884c\u5b89\u88c5\u5e76\u786e\u4fdd\u5176\u5b58\u50a8\u76ee\u7684\u5730\u6709\u8db3\u591f\u5927\u7684\u78c1\u76d8\u7a7a\u95f4\u3002"]}),"\n",(0,r.jsx)(s.h3,{id:"\u90e8\u7f72-ob-operator",children:"\u90e8\u7f72 ob-operator"}),"\n",(0,r.jsx)(s.h4,{id:"\u4f7f\u7528-yaml-\u914d\u7f6e\u6587\u4ef6",children:"\u4f7f\u7528 YAML \u914d\u7f6e\u6587\u4ef6\xb7"}),"\n",(0,r.jsx)(s.p,{children:"\u901a\u8fc7\u4ee5\u4e0b\u547d\u4ee4\u5373\u53ef\u5728 K8s \u96c6\u7fa4\u4e2d\u90e8\u7f72 ob-operator\uff1a"}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsx)(s.li,{children:"\u7a33\u5b9a\u7248\u672c"}),"\n"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.2.0_release/deploy/operator.yaml\n"})}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsx)(s.li,{children:"\u5f00\u53d1\u7248\u672c"}),"\n"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/master/deploy/operator.yaml\n"})}),"\n",(0,r.jsx)(s.h4,{id:"\u4f7f\u7528-helm-chart",children:"\u4f7f\u7528 Helm Chart"}),"\n",(0,r.jsx)(s.p,{children:"Helm Chart \u5c06 ob-operator \u90e8\u7f72\u7684\u547d\u540d\u7a7a\u95f4\u8fdb\u884c\u4e86\u53c2\u6570\u5316\uff0c\u53ef\u5728\u5b89\u88c5 ob-operator \u4e4b\u524d\u6307\u5b9a\u547d\u540d\u7a7a\u95f4\u3002"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"helm repo add ob-operator https://oceanbase.github.io/ob-operator/\nhelm repo update\nhelm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace --version=2.2.0\n"})}),"\n",(0,r.jsx)(s.h4,{id:"\u4f7f\u7528-terraform",children:"\u4f7f\u7528 terraform"}),"\n",(0,r.jsxs)(s.p,{children:["\u90e8\u7f72\u6240\u9700\u8981\u7684\u6587\u4ef6\u653e\u5728\u9879\u76ee\u7684 ",(0,r.jsx)(s.code,{children:"deploy/terraform"})," \u76ee\u5f55"]}),"\n",(0,r.jsxs)(s.ol,{children:["\n",(0,r.jsxs)(s.li,{children:["\u751f\u6210\u914d\u7f6e\u53d8\u91cf:\n\u5728\u5f00\u59cb\u90e8\u7f72\u524d\uff0c\u9700\u8981\u901a\u8fc7\u4ee5\u4e0b\u547d\u4ee4\u6765\u751f\u6210 ",(0,r.jsx)(s.code,{children:"terraform.tfvars"})," \u6587\u4ef6\uff0c\u7528\u6765\u8bb0\u5f55\u5f53\u524d Kubernetes \u96c6\u7fa4\u7684\u4e00\u4e9b\u914d\u7f6e\u3002"]}),"\n"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"cd deploy/terraform\n./generate_k8s_cluster_tfvars.sh\n"})}),"\n",(0,r.jsxs)(s.ol,{start:"2",children:["\n",(0,r.jsx)(s.li,{children:"\u521d\u59cb\u5316 Terraform:\n\u6b64\u6b65\u9aa4\u7528\u6765\u4fdd\u8bc1 terraform \u83b7\u53d6\u5230\u5fc5\u8981\u7684 plugin \u548c\u6a21\u5757\u6765\u7ba1\u7406\u914d\u7f6e\u7684\u8d44\u6e90\uff0c\u4f7f\u7528\u5982\u4e0b\u547d\u4ee4\u6765\u8fdb\u884c\u521d\u59cb\u5316\u3002"}),"\n"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"terraform init\n"})}),"\n",(0,r.jsxs)(s.ol,{start:"3",children:["\n",(0,r.jsx)(s.li,{children:"\u5e94\u7528\u914d\u7f6e:\n\u6267\u884c\u4ee5\u4e0b\u547d\u4ee4\u5f00\u59cb\u90e8\u7f72 ob-operator\u3002"}),"\n"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"terraform apply\n"})}),"\n",(0,r.jsx)(s.h4,{id:"\u9a8c\u8bc1\u90e8\u7f72\u7ed3\u679c",children:"\u9a8c\u8bc1\u90e8\u7f72\u7ed3\u679c"}),"\n",(0,r.jsx)(s.p,{children:"\u5b89\u88c5\u5b8c\u6210\u4e4b\u540e\uff0c\u53ef\u4ee5\u4f7f\u7528\u4ee5\u4e0b\u547d\u4ee4\u9a8c\u8bc1 ob-operator \u662f\u5426\u90e8\u7f72\u6210\u529f\uff1a"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"kubectl get pod -n oceanbase-system\n\n# \u9884\u671f\u7684\u8f93\u51fa\nNAME                                            READY   STATUS    RESTARTS   AGE\noceanbase-controller-manager-86cfc8f7bf-4hfnj   2/2     Running   0          1m\n"})}),"\n",(0,r.jsx)(s.h3,{id:"\u90e8\u7f72-oceanbase-\u96c6\u7fa4",children:"\u90e8\u7f72 OceanBase \u96c6\u7fa4"}),"\n",(0,r.jsx)(s.p,{children:"\u521b\u5efa OceanBase \u96c6\u7fa4\u4e4b\u524d\uff0c\u9700\u8981\u5148\u521b\u5efa\u597d\u82e5\u5e72 secret \u6765\u5b58\u50a8 OceanBase \u4e2d\u7684\u7279\u5b9a\u7528\u6237\u7684\u5bc6\u7801\uff1a"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"kubectl create secret generic root-password --from-literal=password='root_password'\n"})}),"\n",(0,r.jsx)(s.p,{children:"\u901a\u8fc7\u4ee5\u4e0b\u547d\u4ee4\u5373\u53ef\u5728 K8s \u96c6\u7fa4\u4e2d\u90e8\u7f72 OceanBase\uff1a"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.2.0_release/example/quickstart/obcluster.yaml\n"})}),"\n",(0,r.jsx)(s.p,{children:"\u4e00\u822c\u521d\u59cb\u5316\u96c6\u7fa4\u9700\u8981 2 \u5206\u949f\u5de6\u53f3\u7684\u65f6\u95f4\uff0c\u6267\u884c\u4ee5\u4e0b\u547d\u4ee4\uff0c\u67e5\u8be2\u96c6\u7fa4\u72b6\u6001\uff0c\u5f53\u96c6\u7fa4\u72b6\u6001\u53d8\u6210 running \u4e4b\u540e\u8868\u793a\u96c6\u7fa4\u521b\u5efa\u548c\u521d\u59cb\u5316\u6210\u529f\uff1a"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"kubectl get obclusters.oceanbase.oceanbase.com test\n\n# desired output \nNAME   STATUS    AGE\ntest   running   6m2s\n"})}),"\n",(0,r.jsx)(s.h3,{id:"\u8fde\u63a5\u96c6\u7fa4",children:"\u8fde\u63a5\u96c6\u7fa4"}),"\n",(0,r.jsxs)(s.p,{children:["\u901a\u8fc7\u4ee5\u4e0b\u547d\u4ee4\u67e5\u627e observer \u7684 POD IP\uff0cPOD \u540d\u7684\u89c4\u5219\u662f ",(0,r.jsx)(s.code,{children:"${cluster_name}-${cluster_id}-${zone}-uuid"}),"\uff1a"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"kubectl get pods -o wide\n"})}),"\n",(0,r.jsx)(s.p,{children:"\u901a\u8fc7\u4ee5\u4e0b\u547d\u4ee4\u8fde\u63a5\uff1a"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-shell",children:"mysql -h{POD_IP} -P2881 -uroot -proot_password oceanbase -A -c\n"})}),"\n",(0,r.jsx)(s.h3,{id:"oceanbase-dashboard",children:"OceanBase Dashboard"}),"\n",(0,r.jsx)(s.p,{children:"\u6211\u4eec\u5f88\u9ad8\u5174\u5411\u7528\u6237\u63a8\u51fa\u521b\u65b0\u7684 OceanBase Kubernetes Dashboard\uff0c\u8fd9\u662f\u4e00\u6b3e\u65e8\u5728\u6539\u5584\u7528\u6237\u5728 Kubernetes \u4e0a\u7ba1\u7406\u548c\u76d1\u63a7 OceanBase \u96c6\u7fa4\u4f53\u9a8c\u7684\u5148\u8fdb\u5de5\u5177\u3002\u6b22\u8fce\u5404\u4f4d\u7528\u6237\u4f7f\u7528\u548c\u53cd\u9988\uff0c\u540c\u65f6\u6211\u4eec\u4e5f\u5728\u79ef\u6781\u5f00\u53d1\u65b0\u529f\u80fd\u4ee5\u589e\u5f3a\u672a\u6765\u7684\u66f4\u65b0\u3002"}),"\n",(0,r.jsx)(s.p,{children:"\u5b89\u88c5 OceanBase Dashboard \u975e\u5e38\u7b80\u5355, \u53ea\u9700\u8981\u6267\u884c\u5982\u4e0b\u547d\u4ee4\u3002"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"helm repo add ob-operator https://oceanbase.github.io/ob-operator/\nhelm repo update ob-operator\nhelm install oceanbase-dashboard ob-operator/oceanbase-dashboard --version=0.2.0\n"})}),"\n",(0,r.jsx)(s.p,{children:(0,r.jsx)(s.img,{alt:"oceanbase-dashboard-install",src:n(155).A+"",width:"1679",height:"786"})}),"\n",(0,r.jsx)(s.p,{children:"OceanBase Dashboard \u6210\u529f\u5b89\u88c5\u4e4b\u540e, \u4f1a\u81ea\u52a8\u521b\u5efa\u4e00\u4e2a admin \u7528\u6237\u548c\u968f\u673a\u5bc6\u7801\uff0c\u53ef\u4ee5\u901a\u8fc7\u5982\u4e0b\u547d\u4ee4\u67e5\u770b\u5bc6\u7801\u3002"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"echo $(kubectl get -n default secret oceanbase-dashboard-user-credentials -o jsonpath='{.data.admin}' | base64 -d)\n"})}),"\n",(0,r.jsx)(s.p,{children:"\u4e00\u4e2a NodePort \u7c7b\u578b\u7684 service \u4f1a\u9ed8\u8ba4\u521b\u5efa\uff0c\u53ef\u4ee5\u901a\u8fc7\u5982\u4e0b\u547d\u4ee4\u67e5\u770b service \u7684\u5730\u5740\uff0c\u7136\u540e\u5728\u6d4f\u89c8\u5668\u4e2d\u6253\u5f00\u3002"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"kubectl get svc oceanbase-dashboard-ob-dashboard\n"})}),"\n",(0,r.jsx)(s.p,{children:(0,r.jsx)(s.img,{alt:"oceanbase-dashboard-service",src:n(1695).A+"",width:"1478",height:"110"})}),"\n",(0,r.jsxs)(s.p,{children:["\u4f7f\u7528 admin \u8d26\u53f7\u548c\u67e5\u770b\u5230\u7684\u5bc6\u7801\u767b\u5f55\u3002\n",(0,r.jsx)(s.img,{alt:"oceanbase-dashboard-overview",src:n(9133).A+"",width:"3840",height:"2020"})]}),"\n",(0,r.jsx)(s.h2,{id:"\u9879\u76ee\u67b6\u6784",children:"\u9879\u76ee\u67b6\u6784"}),"\n",(0,r.jsx)(s.p,{children:"ob-operator \u4ee5 kubebuilder \u4e3a\u57fa\u7840\uff0c\u901a\u8fc7\u7edf\u4e00\u7684\u8d44\u6e90\u7ba1\u7406\u5668\u63a5\u53e3\u3001\u5168\u5c40\u7684\u4efb\u52a1\u7ba1\u7406\u5668\u5b9e\u4f8b\u4ee5\u53ca\u89e3\u51b3\u957f\u8c03\u5ea6\u7684\u4efb\u52a1\u6d41\u673a\u5236\u5b8c\u6210\u5bf9 OceanBase \u96c6\u7fa4\u53ca\u76f8\u5173\u5e94\u7528\u7684\u63a7\u5236\u548c\u7ba1\u7406\u3002ob-operator \u7684\u67b6\u6784\u5927\u81f4\u5982\u4e0b\u56fe\u6240\u793a\uff1a"}),"\n",(0,r.jsx)(s.p,{children:(0,r.jsx)(s.img,{alt:"ob-operator \u67b6\u6784\u8bbe\u8ba1",src:n(2818).A+"",width:"4079",height:"2117"})}),"\n",(0,r.jsxs)(s.p,{children:["\u6709\u5173\u67b6\u6784\u7ec6\u8282\u53ef\u53c2\u89c1",(0,r.jsx)(l.A,{to:"docs/developer/arch",children:"\u67b6\u6784\u8bbe\u8ba1\u6587\u6863"}),"\u3002"]}),"\n",(0,r.jsx)(s.h2,{id:"\u7279\u6027",children:"\u7279\u6027"}),"\n",(0,r.jsx)(s.p,{children:"ob-operator \u652f\u6301 OceanBase \u96c6\u7fa4\u7684\u7ba1\u7406\u3001\u79df\u6237\u7ba1\u7406\u3001\u5907\u4efd\u6062\u590d\u3001\u6545\u969c\u6062\u590d\u7b49\u529f\u80fd\uff0c\u5177\u4f53\u800c\u8a00\u652f\u6301\u4e86\u4ee5\u4e0b\u529f\u80fd\uff1a"}),"\n",(0,r.jsxs)(s.ul,{className:"contains-task-list",children:["\n",(0,r.jsxs)(s.li,{className:"task-list-item",children:[(0,r.jsx)(s.input,{type:"checkbox",checked:!0,disabled:!0})," ","\u96c6\u7fa4\u7ba1\u7406\uff1a\u96c6\u7fa4\u81ea\u4e3e\u3001\u8c03\u6574\u96c6\u7fa4\u62d3\u6251\u3001\u652f\u6301 K8s \u62d3\u6251\u914d\u7f6e\u3001\u6269\u7f29\u5bb9\u3001\u96c6\u7fa4\u5347\u7ea7\u3001\u4fee\u6539\u53c2\u6570"]}),"\n",(0,r.jsxs)(s.li,{className:"task-list-item",children:[(0,r.jsx)(s.input,{type:"checkbox",checked:!0,disabled:!0})," ","\u79df\u6237\u7ba1\u7406\uff1a\u521b\u5efa\u79df\u6237\u3001\u8c03\u6574\u79df\u6237\u62d3\u6251\u3001\u7ba1\u7406\u8d44\u6e90\u5355\u5143\u3001\u4fee\u6539\u7528\u6237\u5bc6\u7801"]}),"\n",(0,r.jsxs)(s.li,{className:"task-list-item",children:[(0,r.jsx)(s.input,{type:"checkbox",checked:!0,disabled:!0})," ","\u5907\u4efd\u6062\u590d\uff1a\u5411 OSS \u6216 NFS \u76ee\u7684\u5730\u5468\u671f\u6027\u5907\u4efd\u6570\u636e\u3001\u4ece OSS \u6216 NFS \u4e2d\u6062\u590d\u6570\u636e"]}),"\n",(0,r.jsxs)(s.li,{className:"task-list-item",children:[(0,r.jsx)(s.input,{type:"checkbox",checked:!0,disabled:!0})," ","\u7269\u7406\u5907\u5e93\uff1a\u4ece\u5907\u4efd\u4e2d\u6062\u590d\u51fa\u5907\u79df\u6237\u3001\u521b\u5efa\u7a7a\u5907\u79df\u6237\u3001\u5907\u79df\u6237\u5347\u4e3b\u3001\u4e3b\u5907\u5207\u6362"]}),"\n",(0,r.jsxs)(s.li,{className:"task-list-item",children:[(0,r.jsx)(s.input,{type:"checkbox",checked:!0,disabled:!0})," ","\u6545\u969c\u6062\u590d\uff1a\u5355\u8282\u70b9\u6545\u969c\u6062\u590d\uff0cIP \u4fdd\u6301\u60c5\u51b5\u4e0b\u7684\u96c6\u7fa4\u6545\u969c\u6062\u590d"]}),"\n"]}),"\n",(0,r.jsx)(s.p,{children:"\u5373\u5c06\u652f\u6301\u7684\u529f\u80fd\u6709\uff1a"}),"\n",(0,r.jsxs)(s.ul,{className:"contains-task-list",children:["\n",(0,r.jsxs)(s.li,{className:"task-list-item",children:[(0,r.jsx)(s.input,{type:"checkbox",disabled:!0})," ","Dashboard\uff1a\u57fa\u4e8e ob-operator \u7684\u56fe\u5f62\u5316 OceanBase \u96c6\u7fa4\u7ba1\u7406\u5de5\u5177"]}),"\n",(0,r.jsxs)(s.li,{className:"task-list-item",children:[(0,r.jsx)(s.input,{type:"checkbox",disabled:!0})," ","\u4e30\u5bcc\u7684\u8fd0\u7ef4\u4efb\u52a1\u8d44\u6e90\uff1a\u5305\u62ec\u4f46\u4e0d\u9650\u4e8e\u9488\u5bf9\u96c6\u7fa4\u548c\u79df\u6237\u7684\u8f7b\u91cf\u4efb\u52a1"]}),"\n"]}),"\n",(0,r.jsx)(s.h2,{id:"\u652f\u6301\u7684-oceanbase-\u7248\u672c",children:"\u652f\u6301\u7684 OceanBase \u7248\u672c"}),"\n",(0,r.jsx)(s.p,{children:"ob-operator \u652f\u6301 OceanBase v4.x \u7248\u672c\u3002\u67d0\u4e9b\u7279\u6027\u9700\u8981\u7279\u5b9a\u7684 OceanBase \u7248\u672c\uff0c\u53ef\u53c2\u8003\u7528\u6237\u624b\u518c\u83b7\u53d6\u8be6\u7ec6\u4fe1\u606f\u3002"}),"\n",(0,r.jsx)(s.p,{children:"\u6682\u4e0d\u652f\u6301 OceanBase v3.x \u7248\u672c\u3002"}),"\n",(0,r.jsx)(s.h2,{id:"\u73af\u5883\u4f9d\u8d56",children:"\u73af\u5883\u4f9d\u8d56"}),"\n",(0,r.jsxs)(s.p,{children:["ob-operator \u4f7f\u7528 ",(0,r.jsx)(s.a,{href:"https://book.kubebuilder.io/introduction",children:"kubebuilder"})," \u9879\u76ee\u8fdb\u884c\u6784\u5efa\uff0c\u6240\u4ee5\u5f00\u53d1\u548c\u8fd0\u884c\u73af\u5883\u4e0e\u5176\u76f8\u8fd1\u3002"]}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsx)(s.li,{children:"\u6784\u5efa ob-operator \u9700\u8981 Go 1.20 \u7248\u672c\u53ca\u4ee5\u4e0a\uff1b"}),"\n",(0,r.jsx)(s.li,{children:"\u8fd0\u884c ob-operator \u9700\u8981 Kubernetes \u96c6\u7fa4\u548c kubectl \u7684\u7248\u672c\u5728 1.18 \u53ca\u4ee5\u4e0a\u3002\u6211\u4eec\u5728 1.23 ~ 1.25 \u7248\u672c\u7684 K8s \u96c6\u7fa4\u4e0a\u68c0\u9a8c\u8fc7 ob-operator \u7684\u8fd0\u884c\u662f\u7b26\u5408\u9884\u671f\u7684\u3002"}),"\n",(0,r.jsx)(s.li,{children:"\u5982\u679c\u4f7f\u7528 Docker \u4f5c\u4e3a\u96c6\u7fa4\u7684\u5bb9\u5668\u8fd0\u884c\u65f6\uff0c\u9700\u8981 Docker 17.03 \u53ca\u4ee5\u4e0a\u7248\u672c\uff1b\u6211\u4eec\u7684\u6784\u5efa\u548c\u8fd0\u884c\u73af\u5883\u4f7f\u7528\u7684 Docker \u7248\u672c\u4e3a 18\u3002"}),"\n"]}),"\n",(0,r.jsx)(s.h2,{id:"\u6587\u6863",children:"\u6587\u6863"}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsxs)(s.li,{children:["\n",(0,r.jsx)(l.A,{to:"docs/developer/arch",children:"ob-operator \u67b6\u6784\u8bbe\u8ba1"}),"\n"]}),"\n",(0,r.jsxs)(s.li,{children:["\n",(0,r.jsx)(l.A,{to:"docs/developer/deploy",children:"\u90e8\u7f72 ob-operator"}),"\n"]}),"\n",(0,r.jsxs)(s.li,{children:["\n",(0,r.jsx)(l.A,{to:"docs/developer/contributor-guidance",children:"\u5f00\u53d1\u624b\u518c"}),"\n"]}),"\n",(0,r.jsxs)(s.li,{children:["\n",(0,r.jsx)(l.A,{to:"docs/manual/what-is-ob-operator",children:"\u7528\u6237\u624b\u518c"}),"\n"]}),"\n"]}),"\n",(0,r.jsx)(s.h2,{id:"\u83b7\u53d6\u5e2e\u52a9",children:"\u83b7\u53d6\u5e2e\u52a9"}),"\n",(0,r.jsx)(s.p,{children:"\u5982\u679c\u60a8\u5728\u4f7f\u7528 ob-operator \u65f6\u9047\u5230\u4efb\u4f55\u95ee\u9898\uff0c\u6b22\u8fce\u901a\u8fc7\u4ee5\u4e0b\u65b9\u5f0f\u5bfb\u6c42\u5e2e\u52a9\uff1a"}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.a,{href:"https://github.com/oceanbase/ob-operator/issues",children:"GitHub Issue"})}),"\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.a,{href:"https://ask.oceanbase.com/",children:"\u5b98\u65b9\u8bba\u575b"})}),"\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.a,{href:"https://oceanbase.slack.com/archives/C053PT371S7",children:"Slack"})}),"\n",(0,r.jsxs)(s.li,{children:["\u9489\u9489\u7fa4\uff08",(0,r.jsx)(s.a,{target:"_blank","data-noBrokenLinkCheck":!0,href:n(8609).A+"",children:"\u4e8c\u7ef4\u7801"}),"\uff09"]}),"\n",(0,r.jsx)(s.li,{children:"\u5fae\u4fe1\u7fa4\uff08\u8bf7\u6dfb\u52a0\u5c0f\u52a9\u624b\u5fae\u4fe1\uff0c\u5fae\u4fe1\u53f7: OBCE666\uff09"}),"\n"]}),"\n",(0,r.jsx)(s.h2,{id:"\u53c2\u4e0e\u5f00\u53d1",children:"\u53c2\u4e0e\u5f00\u53d1"}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.a,{href:"https://github.com/oceanbase/ob-operator/issues",children:"\u63d0\u51fa Issue"})}),"\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.a,{href:"https://github.com/oceanbase/ob-operator/pulls",children:"\u53d1\u8d77 Pull request"})}),"\n"]}),"\n",(0,r.jsx)(s.h2,{id:"\u8bb8\u53ef\u8bc1",children:"\u8bb8\u53ef\u8bc1"}),"\n",(0,r.jsxs)(s.p,{children:["ob-operator \u4f7f\u7528 ",(0,r.jsx)(s.a,{href:"http://license.coscl.org.cn/MulanPSL2",children:"MulanPSL - 2.0"})," \u8bb8\u53ef\u8bc1\u3002\n\u60a8\u53ef\u4ee5\u514d\u8d39\u590d\u5236\u53ca\u4f7f\u7528\u6e90\u4ee3\u7801\u3002\u5f53\u60a8\u4fee\u6539\u6216\u5206\u53d1\u6e90\u4ee3\u7801\u65f6\uff0c\u8bf7\u9075\u5b88\u6728\u5170\u534f\u8bae\u3002"]})]})}function p(e={}){const{wrapper:s}={...(0,a.R)(),...e.components};return s?(0,r.jsx)(s,{...e,children:(0,r.jsx)(h,{...e})}):h(e)}},8609:(e,s,n)=>{n.d(s,{A:()=>r});const r=n.p+"assets/files/dingtalk-operator-users-6feeeda042ca96cb0d62dc20144f88c2.png"},2818:(e,s,n)=>{n.d(s,{A:()=>r});const r=n.p+"assets/images/ob-operator-arch-af746e5c9ef3dc1c9ce493fc38b54820.png"},155:(e,s,n)=>{n.d(s,{A:()=>r});const r=n.p+"assets/images/oceanbase-dashboard-install-cae3cd61ca913319c39b7ae2519c2af7.jpg"},9133:(e,s,n)=>{n.d(s,{A:()=>r});const r=n.p+"assets/images/oceanbase-dashboard-overview-11f1acf292b32358be4c85b8e0e9f41c.jpg"},1695:(e,s,n)=>{n.d(s,{A:()=>r});const r=n.p+"assets/images/oceanbase-dashboard-service-c10652b2c946be814ab204ef049efad0.jpg"},8453:(e,s,n)=>{n.d(s,{R:()=>c,x:()=>o});var r=n(6540);const a={},l=r.createContext(a);function c(e){const s=r.useContext(l);return r.useMemo((function(){return"function"==typeof e?e(s):{...s,...e}}),[s,e])}function o(e){let s;return s=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:c(e.components),r.createElement(l.Provider,{value:s},e.children)}}}]);
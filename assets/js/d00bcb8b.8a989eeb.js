"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[512],{8548:(e,o,n)=>{n.r(o),n.d(o,{assets:()=>c,contentTitle:()=>r,default:()=>p,frontMatter:()=>l,metadata:()=>a,toc:()=>i});var t=n(4848),s=n(8453);const l={},r="Deploy ob-operator",a={id:"developer/deploy",title:"Deploy ob-operator",description:"This article introduces the deployment methods for ob-operator.",source:"@site/docs/developer/deploy.md",sourceDirName:"developer",slug:"/developer/deploy",permalink:"/ob-operator/docs/developer/deploy",draft:!1,unlisted:!1,editUrl:"https://github.com/oceanbase/ob-operator/tree/master/docsite/docs/developer/deploy.md",tags:[],version:"current",frontMatter:{},sidebar:"developerSidebar",previous:{title:"Deploy ob-operator locally",permalink:"/ob-operator/docs/developer/deploy-locally"},next:{title:"Develop ob-operator locally",permalink:"/ob-operator/docs/developer/develop-locally"}},c={},i=[{value:"1. Deployment Dependencies",id:"1-deployment-dependencies",level:2},{value:"2.1 Deploying with Helm",id:"21-deploying-with-helm",level:2},{value:"2.2 Deploying with Configuration Files",id:"22-deploying-with-configuration-files",level:2},{value:"3. Check the deployment results",id:"3-check-the-deployment-results",level:2}];function d(e){const o={a:"a",code:"code",h1:"h1",h2:"h2",li:"li",p:"p",pre:"pre",ul:"ul",...(0,s.R)(),...e.components};return(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(o.h1,{id:"deploy-ob-operator",children:"Deploy ob-operator"}),"\n",(0,t.jsx)(o.p,{children:"This article introduces the deployment methods for ob-operator."}),"\n",(0,t.jsx)(o.h2,{id:"1-deployment-dependencies",children:"1. Deployment Dependencies"}),"\n",(0,t.jsxs)(o.p,{children:["ob-operator relies on ",(0,t.jsx)(o.a,{href:"https://cert-manager.io/docs/",children:"cert-manager"}),". You can refer to the corresponding installation documentation for the ",(0,t.jsx)(o.a,{href:"https://cert-manager.io/docs/installation/",children:"installation of cert-manager"}),"."]}),"\n",(0,t.jsx)(o.h2,{id:"21-deploying-with-helm",children:"2.1 Deploying with Helm"}),"\n",(0,t.jsxs)(o.p,{children:["ob-operator supports deployment using Helm. Before deploying ob-operator with the Helm command, you need to install ",(0,t.jsx)(o.a,{href:"https://github.com/helm/helm",children:"Helm"}),". After Helm is installed, you can deploy ob-operator directly using the following command."]}),"\n",(0,t.jsx)(o.pre,{children:(0,t.jsx)(o.code,{className:"language-shell",children:"helm repo add ob-operator https://oceanbase.github.io/ob-operator/\nhelm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace --version=2.2.0\n"})}),"\n",(0,t.jsx)(o.p,{children:"Parameters:"}),"\n",(0,t.jsxs)(o.ul,{children:["\n",(0,t.jsxs)(o.li,{children:["\n",(0,t.jsx)(o.p,{children:'namespace: Namespace, can be customized. It is recommended to use "oceanbase-system" as the namespace.'}),"\n"]}),"\n",(0,t.jsxs)(o.li,{children:["\n",(0,t.jsxs)(o.p,{children:["version: ob-operator version number. It is recommended to use the latest version ",(0,t.jsx)(o.code,{children:"2.2.0"}),"."]}),"\n"]}),"\n"]}),"\n",(0,t.jsx)(o.h2,{id:"22-deploying-with-configuration-files",children:"2.2 Deploying with Configuration Files"}),"\n",(0,t.jsxs)(o.ul,{children:["\n",(0,t.jsx)(o.li,{children:"Stable"}),"\n"]}),"\n",(0,t.jsx)(o.pre,{children:(0,t.jsx)(o.code,{className:"language-shell",children:"kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.2.0_release/deploy/operator.yaml\n"})}),"\n",(0,t.jsxs)(o.ul,{children:["\n",(0,t.jsx)(o.li,{children:"Development"}),"\n"]}),"\n",(0,t.jsx)(o.pre,{children:(0,t.jsx)(o.code,{className:"language-shell",children:"kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/master/deploy/operator.yaml\n"})}),"\n",(0,t.jsx)(o.p,{children:"It is generally recommended to use the configuration files for the stable version. However, if you want to use a development version, you can choose to use the configuration files for the development version."}),"\n",(0,t.jsx)(o.h2,{id:"3-check-the-deployment-results",children:"3. Check the deployment results"}),"\n",(0,t.jsx)(o.p,{children:"After a successful deployment, you can view the definition of Custom Resource Definitions (CRDs) by executing the following command:"}),"\n",(0,t.jsx)(o.pre,{children:(0,t.jsx)(o.code,{className:"language-shell",children:"kubectl get crds\n"})}),"\n",(0,t.jsx)(o.p,{children:"If you get the following output, it indicates a successful deployment:"}),"\n",(0,t.jsx)(o.pre,{children:(0,t.jsx)(o.code,{className:"language-shell",children:"obparameters.oceanbase.oceanbase.com             2023-11-12T08:06:58Z\nobservers.oceanbase.oceanbase.com                2023-11-12T08:06:58Z\nobtenantbackups.oceanbase.oceanbase.com          2023-11-12T08:06:58Z\nobtenantrestores.oceanbase.oceanbase.com         2023-11-12T08:06:58Z\nobzones.oceanbase.oceanbase.com                  2023-11-12T08:06:58Z\nobtenants.oceanbase.oceanbase.com                2023-11-12T08:06:58Z\nobtenantoperations.oceanbase.oceanbase.com       2023-11-12T08:06:58Z\nobclusters.oceanbase.oceanbase.com               2023-11-12T08:06:58Z\nobtenantbackuppolicies.oceanbase.oceanbase.com   2023-11-12T08:06:58Z\n"})}),"\n",(0,t.jsx)(o.p,{children:"To confirm whether ob-operator has been successfully deployed, you can use the following command:"}),"\n",(0,t.jsx)(o.pre,{children:(0,t.jsx)(o.code,{className:"language-shell",children:"kubectl get pods -n oceanbase-system\n"})}),"\n",(0,t.jsx)(o.p,{children:'The result will look like the following example. If you see that all containers are ready and the status is "Running", it indicates a successful deployment.'}),"\n",(0,t.jsx)(o.pre,{children:(0,t.jsx)(o.code,{className:"language-shell",children:"NAME                                            READY   STATUS    RESTARTS   AGE\noceanbase-controller-manager-86cfc8f7bf-4hfnj   2/2     Running   0          1m\n"})})]})}function p(e={}){const{wrapper:o}={...(0,s.R)(),...e.components};return o?(0,t.jsx)(o,{...e,children:(0,t.jsx)(d,{...e})}):d(e)}},8453:(e,o,n)=>{n.d(o,{R:()=>r,x:()=>a});var t=n(6540);const s={},l=t.createContext(s);function r(e){const o=t.useContext(l);return t.useMemo((function(){return"function"==typeof e?e(o):{...o,...e}}),[o,e])}function a(e){let o;return o=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:r(e.components),t.createElement(l.Provider,{value:o},e.children)}}}]);
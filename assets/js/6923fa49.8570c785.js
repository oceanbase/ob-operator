"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[575],{6393:(e,t,r)=>{r.r(t),r.d(t,{assets:()=>i,contentTitle:()=>s,default:()=>d,frontMatter:()=>n,metadata:()=>c,toc:()=>u});var a=r(4848),o=r(8453);const n={sidebar_position:6},s="Manage parameters",c={id:"manual/ob-operator-user-guide/cluster-management-of-ob-operator/parameter-management",title:"Manage parameters",description:"This topic describes how to modify the parameters of an OceanBase cluster by using ob-operator.",source:"@site/docs/manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator/600.parameter-management.md",sourceDirName:"manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator",slug:"/manual/ob-operator-user-guide/cluster-management-of-ob-operator/parameter-management",permalink:"/ob-operator/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/parameter-management",draft:!1,unlisted:!1,editUrl:"https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/docs/manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator/600.parameter-management.md",tags:[],version:"current",sidebarPosition:6,frontMatter:{sidebar_position:6},sidebar:"manualSidebar",previous:{title:"Upgrade a cluster",permalink:"/ob-operator/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/upgrade-cluster-of-ob-operator"},next:{title:"Update resources",permalink:"/ob-operator/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/update-resources"}},i={},u=[{value:"Prerequisites",id:"prerequisites",level:2},{value:"Procedure",id:"procedure",level:2},{value:"Modify the tag setting in <code>spec</code>",id:"modify-the-tag-setting-in-spec",level:3}];function l(e){const t={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",ol:"ol",p:"p",pre:"pre",...(0,o.R)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsx)(t.h1,{id:"manage-parameters",children:"Manage parameters"}),"\n",(0,a.jsx)(t.p,{children:"This topic describes how to modify the parameters of an OceanBase cluster by using ob-operator."}),"\n",(0,a.jsx)(t.h2,{id:"prerequisites",children:"Prerequisites"}),"\n",(0,a.jsxs)(t.p,{children:["Make sure that the OceanBase cluster is in the ",(0,a.jsx)(t.code,{children:"Running"})," state."]}),"\n",(0,a.jsx)(t.h2,{id:"procedure",children:"Procedure"}),"\n",(0,a.jsxs)(t.h3,{id:"modify-the-tag-setting-in-spec",children:["Modify the tag setting in ",(0,a.jsx)(t.code,{children:"spec"})]}),"\n",(0,a.jsxs)(t.ol,{children:["\n",(0,a.jsxs)(t.li,{children:["\n",(0,a.jsxs)(t.p,{children:["Modify the configuration file of the OceanBase cluster. You need to specify the parameter that you want to modify in ",(0,a.jsx)(t.code,{children:"spec.parameters"}),". For more information about the complete configuration file, see ",(0,a.jsx)(t.a,{href:"/ob-operator/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/create-cluster",children:"Create a cluster"}),"."]}),"\n",(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-yaml",children:"# Before modification\n# parameters:\n#   - name: system_memory\n#     value: 2G\n\n# After modification\nparameters:\n    - name: system_memory\n    value: 2G\n"})}),"\n"]}),"\n",(0,a.jsxs)(t.li,{children:["\n",(0,a.jsx)(t.p,{children:"Run the following command for the modification to take effect:"}),"\n",(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-yaml",children:"kubectl apply -f obcluster.yaml\n"})}),"\n"]}),"\n",(0,a.jsxs)(t.li,{children:["\n",(0,a.jsx)(t.p,{children:"Query the status of custom resources in the OceanBase cluster to check whether the operation succeeds.\nRun the following command to query the status of the OceanBase cluster."}),"\n",(0,a.jsx)(t.pre,{children:(0,a.jsx)(t.code,{className:"language-shell",children:"kubectl get obclusters.oceanbase.oceanbase.com test -n oceanbase -o yaml\n\n# desired output, only displays status and one result here\nstatus:\nparameter:\n- name: system_memory\nserver: 10.42.0.232:2882\nvalue: 2G\nzone: zone1\nstatus: matched\n"})}),"\n"]}),"\n"]})]})}function d(e={}){const{wrapper:t}={...(0,o.R)(),...e.components};return t?(0,a.jsx)(t,{...e,children:(0,a.jsx)(l,{...e})}):l(e)}},8453:(e,t,r)=>{r.d(t,{R:()=>s,x:()=>c});var a=r(6540);const o={},n=a.createContext(o);function s(e){const t=a.useContext(n);return a.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function c(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:s(e.components),a.createElement(n.Provider,{value:t},e.children)}}}]);
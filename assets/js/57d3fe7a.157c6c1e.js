"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[1423],{840:(e,t,s)=>{s.r(t),s.d(t,{assets:()=>c,contentTitle:()=>a,default:()=>h,frontMatter:()=>r,metadata:()=>i,toc:()=>d});var n=s(4848),o=s(8453);const r={sidebar_position:2},a="FAQ",i={id:"manual/appendix/FAQ",title:"FAQ",description:"1. How do I make sure that a resource is ready?",source:"@site/docs/manual/900.appendix/200.FAQ.md",sourceDirName:"manual/900.appendix",slug:"/manual/appendix/FAQ",permalink:"/ob-operator/docs/manual/appendix/FAQ",draft:!1,unlisted:!1,editUrl:"https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/docs/manual/900.appendix/200.FAQ.md",tags:[],version:"current",sidebarPosition:2,frontMatter:{sidebar_position:2},sidebar:"manualSidebar",previous:{title:"A real-world example",permalink:"/ob-operator/docs/manual/appendix/example"}},c={},d=[{value:"1. How do I make sure that a resource is ready?",id:"1-how-do-i-make-sure-that-a-resource-is-ready",level:2},{value:"2. How do I view the O&amp;M status of a resource?",id:"2-how-do-i-view-the-om-status-of-a-resource",level:2},{value:"3. How do I do troubleshooting for ob-operator and OceanBase?",id:"3-how-do-i-do-troubleshooting-for-ob-operator-and-oceanbase",level:2},{value:"4. How do I fix a &quot;stuck&quot; resource in ob-operator?",id:"4-how-do-i-fix-a-stuck-resource-in-ob-operator",level:2},{value:"Reset",id:"reset",level:3},{value:"Delete",id:"delete",level:3},{value:"Retry",id:"retry",level:3},{value:"Skip",id:"skip",level:3}];function l(e){const t={code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",ul:"ul",...(0,o.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(t.h1,{id:"faq",children:"FAQ"}),"\n",(0,n.jsx)(t.h2,{id:"1-how-do-i-make-sure-that-a-resource-is-ready",children:"1. How do I make sure that a resource is ready?"}),"\n",(0,n.jsx)(t.p,{children:"Assume that you want to view the resource status of a cluster. Run the following command:"}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-shell",children:"kubectl get obclusters.oceanbase.oceanbase.com test -n oceanbase\n"})}),"\n",(0,n.jsxs)(t.p,{children:["If the status is ",(0,n.jsx)(t.code,{children:"running"})," in the response, the resource is ready."]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-shell",children:"# desired output\nNAME   STATUS    AGE\ntest   running   6m2s\n"})}),"\n",(0,n.jsx)(t.h2,{id:"2-how-do-i-view-the-om-status-of-a-resource",children:"2. How do I view the O&M status of a resource?"}),"\n",(0,n.jsx)(t.p,{children:"Assume that you want to view the resource status of a cluster. Run the following command:"}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-shell",children:"kubectl get obclusters.oceanbase.oceanbase.com test -n oceanbase -o yaml\n"})}),"\n",(0,n.jsxs)(t.p,{children:["You can check the status and progress of O&M tasks based on values of parameters in the ",(0,n.jsx)(t.code,{children:"operationContext"})," section in the response."]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-shell",children:"status:\n  image: oceanbase/oceanbase-cloud-native:4.2.0.0-101000032023091319\n  obzones:\n  - status: delete observer\n    zone: obcluster-1-zone1\n  - status: delete observer\n    zone: obcluster-1-zone2\n  - status: delete observer\n    zone: obcluster-1-zone3\n  operationContext:\n    failureRule:\n      failureStatus: running\n      failureStrategy: retry over\n      retryCount: 0\n    idx: 2\n    name: modify obzone replica\n    targetStatus: running\n    task: wait obzone topology match\n    taskId: c04aeb28-01e7-4f85-b390-8d855b9f30e3\n    taskStatus: running\n    tasks:\n    - modify obzone replica\n    - wait obzone topology match\n    - wait obzone running\n  parameters: []\n  status: modify obzone replica\n"})}),"\n",(0,n.jsx)(t.h2,{id:"3-how-do-i-do-troubleshooting-for-ob-operator-and-oceanbase",children:"3. How do I do troubleshooting for ob-operator and OceanBase?"}),"\n",(0,n.jsxs)(t.ul,{children:["\n",(0,n.jsx)(t.li,{children:"Generally, you need to first analyze the logs of ob-operator to locate an error. Run the following command to view the logs of ob-operator:"}),"\n"]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-shell",children:"kubectl logs oceanbase-controller-manager-86cfc8f7bf-js95z -n oceanbase-system -c manager  | less\n"})}),"\n",(0,n.jsxs)(t.ul,{children:["\n",(0,n.jsx)(t.li,{children:"View the logs of the OBServer node"}),"\n"]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-shell",children:"# Log on to the container of the OBServer node.\nkubectl exec -it obcluster-1-zone1-8ab645f4d0f9 -n oceanbase -c observer -- bash\n\n# The directory where the log files are located.\ncd /home/admin/oceanbase/log\n"})}),"\n",(0,n.jsx)(t.h2,{id:"4-how-do-i-fix-a-stuck-resource-in-ob-operator",children:'4. How do I fix a "stuck" resource in ob-operator?'}),"\n",(0,n.jsxs)(t.p,{children:["As ob-operator uses a state machine and task flow to manage custom resources (CRs) and their O&M operations, there may be situations where CRs are in an unexpected state. This could include continuously retrying a task flow that is bound to fail, failing to delete a resource, or mistakenly deleting a resource that needs to be recovered. In cases where a CR cannot be restored to normal through regular operations, you can use the ",(0,n.jsx)(t.code,{children:"OBResourceRescue"})," resource to rescue the problematic CR. The ",(0,n.jsx)(t.code,{children:"OBResourceRescue"})," resource includes four types of operations: ",(0,n.jsx)(t.code,{children:"reset"}),", ",(0,n.jsx)(t.code,{children:"delete"}),", ",(0,n.jsx)(t.code,{children:"retry"}),", and ",(0,n.jsx)(t.code,{children:"skip"}),"."]}),"\n",(0,n.jsxs)(t.p,{children:["A typical ",(0,n.jsx)(t.code,{children:"OBResourceRescue"})," CR configuration is as follows:"]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBResourceRescue\nmetadata:\n  generateName: rescue-reset- # generateName needs to be used with kubectl create -f\nspec:\n  type: reset\n  targetKind: OBCluster\n  targetResName: test\n  targetStatus: running # The target status needs to be filled in when the type is reset\n"})}),"\n",(0,n.jsx)(t.p,{children:"The key configurations are explained in the following table:"}),"\n",(0,n.jsxs)(t.table,{children:[(0,n.jsx)(t.thead,{children:(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.th,{children:"Configuration item"}),(0,n.jsx)(t.th,{children:"Optional values"}),(0,n.jsx)(t.th,{children:"Description"})]})}),(0,n.jsxs)(t.tbody,{children:[(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:"type"}),(0,n.jsxs)(t.td,{children:[(0,n.jsx)(t.code,{children:"reset"}),", ",(0,n.jsx)(t.code,{children:"delete"}),", ",(0,n.jsx)(t.code,{children:"retry"}),", ",(0,n.jsx)(t.code,{children:"skip"})]}),(0,n.jsx)(t.td,{children:"The type of the resource rescue action"})]}),(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:"targetKind"}),(0,n.jsxs)(t.td,{children:[(0,n.jsx)(t.code,{children:"OBCluster"}),", ",(0,n.jsx)(t.code,{children:"OBZone"}),", ",(0,n.jsx)(t.code,{children:"OBTenant"}),", and other CRD kind managed by ob-operator"]}),(0,n.jsx)(t.td,{children:"The kind of the resource to be rescued"})]}),(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:"targetResName"}),(0,n.jsx)(t.td,{children:"/"}),(0,n.jsx)(t.td,{children:"The name of the resource to be rescued"})]}),(0,n.jsxs)(t.tr,{children:[(0,n.jsx)(t.td,{children:"targetStatus"}),(0,n.jsx)(t.td,{children:"/"}),(0,n.jsx)(t.td,{children:"This field needs to be filled in when the type is reset, indicating the status of the resource after the reset"})]})]})]}),"\n",(0,n.jsx)(t.h3,{id:"reset",children:"Reset"}),"\n",(0,n.jsxs)(t.p,{children:["The configuration example of the typical CR above is a reset type of resource rescue. After creating this resource in the K8s cluster using the ",(0,n.jsx)(t.code,{children:"kubectl create -f"})," command, ob-operator sets the ",(0,n.jsx)(t.code,{children:"status.status"})," of the resource whose kind is ",(0,n.jsx)(t.code,{children:"OBCluster"})," and name is ",(0,n.jsx)(t.code,{children:"test"})," to ",(0,n.jsx)(t.code,{children:"running"})," (the ",(0,n.jsx)(t.code,{children:"targetStatus"})," set in the configuration file), and sets the ",(0,n.jsx)(t.code,{children:"status.operationContext"})," of the resource to empty."]}),"\n",(0,n.jsx)(t.h3,{id:"delete",children:"Delete"}),"\n",(0,n.jsxs)(t.p,{children:["The configuration example of the delete type of rescue action is as follows. After creating this resource in the cluster, ob-operator clears the ",(0,n.jsx)(t.code,{children:"finalizers"})," field of the target resource and sets the ",(0,n.jsx)(t.code,{children:"deletionTimestamp"})," of the resource to the current time."]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-yaml",children:"# ...\nspec:\n  type: delete\n  targetKind: OBCluster\n  targetResName: test\n"})}),"\n",(0,n.jsx)(t.h3,{id:"retry",children:"Retry"}),"\n",(0,n.jsxs)(t.p,{children:["The configuration example of the retry type of rescue action is as follows. After creating this resource in the cluster, ob-operator sets the ",(0,n.jsx)(t.code,{children:"status.operationContext.retryCount"})," of the target resource to 0 and sets the ",(0,n.jsx)(t.code,{children:"status.operationContext.taskStatus"})," to ",(0,n.jsx)(t.code,{children:"pending"}),". Resources in this state will retry the current task."]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-yaml",children:"# ...\nspec:\n  type: retry\n  targetKind: OBCluster\n  targetResName: test\n"})}),"\n",(0,n.jsx)(t.h3,{id:"skip",children:"Skip"}),"\n",(0,n.jsxs)(t.p,{children:["The configuration example of the skip type of rescue action is as follows. After creating this resource in the cluster, ob-operator directly sets the ",(0,n.jsx)(t.code,{children:"status.operationContext.taskStatus"})," of the target resource to ",(0,n.jsx)(t.code,{children:"successful"}),". After receiving this message, the task manager will execute the next task in the ",(0,n.jsx)(t.code,{children:"tasks"})," field."]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-yaml",children:"# ...\nspec:\n  type: skip\n  targetKind: OBCluster\n  targetResName: test\n"})})]})}function h(e={}){const{wrapper:t}={...(0,o.R)(),...e.components};return t?(0,n.jsx)(t,{...e,children:(0,n.jsx)(l,{...e})}):l(e)}},8453:(e,t,s)=>{s.d(t,{R:()=>a,x:()=>i});var n=s(6540);const o={},r=n.createContext(o);function a(e){const t=n.useContext(r);return n.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function i(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:a(e.components),n.createElement(r.Provider,{value:t},e.children)}}}]);
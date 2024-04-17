"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[3816],{1524:(e,n,o)=>{o.r(n),o.d(n,{assets:()=>u,contentTitle:()=>t,default:()=>d,frontMatter:()=>s,metadata:()=>c,toc:()=>l});var r=o(4848),a=o(8453);const s={sidebar_position:5},t="\u96c6\u7fa4\u5347\u7ea7",c={id:"manual/ob-operator-user-guide/cluster-management-of-ob-operator/upgrade-cluster-of-ob-operator",title:"\u96c6\u7fa4\u5347\u7ea7",description:"\u672c\u6587\u4ecb\u7ecd\u5347\u7ea7\u4f7f\u7528 ob-operator \u90e8\u7f72\u7684 OceanBase \u96c6\u7fa4\u3002",source:"@site/i18n/zh-Hans/docusaurus-plugin-content-docs/current/manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator/500.upgrade-cluster-of-ob-operator.md",sourceDirName:"manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator",slug:"/manual/ob-operator-user-guide/cluster-management-of-ob-operator/upgrade-cluster-of-ob-operator",permalink:"/ob-operator/zh-Hans/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/upgrade-cluster-of-ob-operator",draft:!1,unlisted:!1,editUrl:"https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/docs/manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator/500.upgrade-cluster-of-ob-operator.md",tags:[],version:"current",sidebarPosition:5,frontMatter:{sidebar_position:5},sidebar:"manualSidebar",previous:{title:"\u4ece Zone \u4e2d\u51cf\u5c11 OBServer \u8282\u70b9",permalink:"/ob-operator/zh-Hans/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/server-management/delete-server"},next:{title:"\u53c2\u6570\u7ba1\u7406",permalink:"/ob-operator/zh-Hans/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/parameter-management"}},u={},l=[{value:"\u524d\u63d0\u6761\u4ef6",id:"\u524d\u63d0\u6761\u4ef6",level:2},{value:"\u64cd\u4f5c\u6b65\u9aa4",id:"\u64cd\u4f5c\u6b65\u9aa4",level:2},{value:"\u4fee\u6539 spec \u4e2d\u7684 tag \u914d\u7f6e",id:"\u4fee\u6539-spec-\u4e2d\u7684-tag-\u914d\u7f6e",level:3}];function i(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",ol:"ol",p:"p",pre:"pre",...(0,a.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(n.h1,{id:"\u96c6\u7fa4\u5347\u7ea7",children:"\u96c6\u7fa4\u5347\u7ea7"}),"\n",(0,r.jsx)(n.p,{children:"\u672c\u6587\u4ecb\u7ecd\u5347\u7ea7\u4f7f\u7528 ob-operator \u90e8\u7f72\u7684 OceanBase \u96c6\u7fa4\u3002"}),"\n",(0,r.jsx)(n.h2,{id:"\u524d\u63d0\u6761\u4ef6",children:"\u524d\u63d0\u6761\u4ef6"}),"\n",(0,r.jsx)(n.p,{children:"\u5728\u96c6\u7fa4\u5347\u7ea7\u524d\uff0c\u60a8\u8981\u786e\u4fdd\u5f85\u5347\u7ea7\u7684\u96c6\u7fa4\u662f running \u72b6\u6001\u3002"}),"\n",(0,r.jsx)(n.h2,{id:"\u64cd\u4f5c\u6b65\u9aa4",children:"\u64cd\u4f5c\u6b65\u9aa4"}),"\n",(0,r.jsx)(n.h3,{id:"\u4fee\u6539-spec-\u4e2d\u7684-tag-\u914d\u7f6e",children:"\u4fee\u6539 spec \u4e2d\u7684 tag \u914d\u7f6e"}),"\n",(0,r.jsxs)(n.ol,{children:["\n",(0,r.jsxs)(n.li,{children:["\n",(0,r.jsxs)(n.p,{children:["\u4fee\u6539 obcluster \u7684\u914d\u7f6e\u6587\u4ef6\u3002\u5b8c\u6574\u914d\u7f6e\u6587\u4ef6\u8bf7\u53c2\u8003 ",(0,r.jsx)(n.a,{href:"/ob-operator/zh-Hans/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/create-cluster",children:"\u521b\u5efa OceanBase \u96c6\u7fa4"}),"\u3002 \u5c06 ",(0,r.jsx)(n.code,{children:"spec.observer.image"})," \u4fee\u6539\u4e3a\u76ee\u6807\u955c\u50cf\u3002"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-yaml",children:"# \u4fee\u6539\u524d\nspec:\n  observer:\n    image: oceanbase/oceanbase-cloud-native:4.2.0.0-101000032023091319\n\n# \u4fee\u6539\u540e\nspec:\n  observer:\n    image: oceanbase/oceanbase-cloud-native:4.2.1.1-101000062023110109\n"})}),"\n"]}),"\n",(0,r.jsxs)(n.li,{children:["\n",(0,r.jsx)(n.p,{children:"\u914d\u7f6e\u6587\u4ef6\u4fee\u6539\u540e\uff0c\u9700\u8fd0\u884c\u5982\u4e0b\u547d\u4ee4\u4f7f\u6539\u52a8\u751f\u6548\u3002"}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-yaml",children:"kubectl apply -f obcluster.yaml\n"})}),"\n"]}),"\n",(0,r.jsxs)(n.li,{children:["\n",(0,r.jsx)(n.p,{children:"\u89c2\u5bdf OceanBase \u96c6\u7fa4 CR \u7684\u72b6\u6001\u7b49\u5f85\u64cd\u4f5c\u6210\u529f\u3002\n\u901a\u8fc7\u4ee5\u4e0b\u547d\u4ee4\uff0c\u53ef\u4ee5\u83b7\u53d6 OceanBase \u96c6\u7fa4\u8d44\u6e90\u7684\u72b6\u6001\uff0c\u5f53\u96c6\u7fa4\u72b6\u6001\u53d8\u4e3a running\uff0cimage \u53d8\u4e3a\u76ee\u6807\u955c\u50cf\u65f6\uff0c\u5219\u5347\u7ea7\u6210\u529f\u3002"}),"\n"]}),"\n"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"kubectl get obclusters.oceanbase.oceanbase.com test -n oceanbase -o yaml\n\n# desired output, only displays status here\nstatus:\n  image: oceanbase/oceanbase-cloud-native:4.2.1.1-101000062023110109\n  obzones:\n  - status: running\n    zone: obcluster-1-zone1\n  - status: running\n    zone: obcluster-1-zone2\n  - status: running\n    zone: obcluster-1-zone3\n  parameters: []\n  status: running\n"})})]})}function d(e={}){const{wrapper:n}={...(0,a.R)(),...e.components};return n?(0,r.jsx)(n,{...e,children:(0,r.jsx)(i,{...e})}):i(e)}},8453:(e,n,o)=>{o.d(n,{R:()=>t,x:()=>c});var r=o(6540);const a={},s=r.createContext(a);function t(e){const n=r.useContext(s);return r.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function c(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:t(e.components),r.createElement(s.Provider,{value:n},e.children)}}}]);
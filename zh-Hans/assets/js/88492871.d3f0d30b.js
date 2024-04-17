"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[6224],{9397:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>d,contentTitle:()=>a,default:()=>h,frontMatter:()=>r,metadata:()=>o,toc:()=>i});var l=n(4848),s=n(8453);const r={sidebar_position:5},a="\u914d\u7f6e ob-operator",o={id:"manual/configuration-of-ob-operator",title:"\u914d\u7f6e ob-operator",description:"\u672c\u6587\u4ecb\u7ecd ob-operator \u7684\u542f\u52a8\u53c2\u6570\u548c\u4f7f\u7528\u5230\u7684\u73af\u5883\u53d8\u91cf\u4ee5\u53ca\u4fee\u6539\u65b9\u6cd5\u3002\u7528\u6237\u53ef\u901a\u8fc7\u6539\u53d8\u542f\u52a8\u53c2\u6570\u548c\u73af\u5883\u53d8\u91cf\u5f71\u54cd ob-operator \u7684\u884c\u4e3a\u3002",source:"@site/i18n/zh-Hans/docusaurus-plugin-content-docs/current/manual/500.configuration-of-ob-operator.md",sourceDirName:"manual",slug:"/manual/configuration-of-ob-operator",permalink:"/ob-operator/zh-Hans/docs/manual/configuration-of-ob-operator",draft:!1,unlisted:!1,editUrl:"https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/docs/manual/500.configuration-of-ob-operator.md",tags:[],version:"current",sidebarPosition:5,frontMatter:{sidebar_position:5},sidebar:"manualSidebar",previous:{title:"ob-operator \u5347\u7ea7",permalink:"/ob-operator/zh-Hans/docs/manual/ob-operator-upgrade"},next:{title:"Manage resources",permalink:"/ob-operator/zh-Hans/docs/category/manage-resources"}},d={},i=[{value:"\u542f\u52a8\u53c2\u6570",id:"\u542f\u52a8\u53c2\u6570",level:2},{value:"\u73af\u5883\u53d8\u91cf",id:"\u73af\u5883\u53d8\u91cf",level:2},{value:"\u4fee\u6539\u65b9\u6cd5",id:"\u4fee\u6539\u65b9\u6cd5",level:2},{value:"\u793a\u4f8b\uff1a\u589e\u5927\u65e5\u5fd7\u8f93\u51fa\u91cf",id:"\u793a\u4f8b\u589e\u5927\u65e5\u5fd7\u8f93\u51fa\u91cf",level:3},{value:"\u793a\u4f8b\uff1a\u6307\u5b9a\u8d44\u6e90\u547d\u540d\u7a7a\u95f4",id:"\u793a\u4f8b\u6307\u5b9a\u8d44\u6e90\u547d\u540d\u7a7a\u95f4",level:3},{value:"\u5e94\u7528\u5230\u96c6\u7fa4\u4e2d",id:"\u5e94\u7528\u5230\u96c6\u7fa4\u4e2d",level:3}];function c(e){const t={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",p:"p",pre:"pre",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,s.R)(),...e.components};return(0,l.jsxs)(l.Fragment,{children:[(0,l.jsx)(t.h1,{id:"\u914d\u7f6e-ob-operator",children:"\u914d\u7f6e ob-operator"}),"\n",(0,l.jsx)(t.p,{children:"\u672c\u6587\u4ecb\u7ecd ob-operator \u7684\u542f\u52a8\u53c2\u6570\u548c\u4f7f\u7528\u5230\u7684\u73af\u5883\u53d8\u91cf\u4ee5\u53ca\u4fee\u6539\u65b9\u6cd5\u3002\u7528\u6237\u53ef\u901a\u8fc7\u6539\u53d8\u542f\u52a8\u53c2\u6570\u548c\u73af\u5883\u53d8\u91cf\u5f71\u54cd ob-operator \u7684\u884c\u4e3a\u3002"}),"\n",(0,l.jsx)(t.h2,{id:"\u542f\u52a8\u53c2\u6570",children:"\u542f\u52a8\u53c2\u6570"}),"\n",(0,l.jsxs)(t.table,{children:[(0,l.jsx)(t.thead,{children:(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.th,{style:{textAlign:"left"},children:"\u53c2\u6570\u540d"}),(0,l.jsx)(t.th,{style:{textAlign:"left"},children:"\u542b\u4e49"}),(0,l.jsx)(t.th,{style:{textAlign:"left"},children:"\u9ed8\u8ba4\u503c"}),(0,l.jsx)(t.th,{style:{textAlign:"left"},children:"deploy \u914d\u7f6e"})]})}),(0,l.jsxs)(t.tbody,{children:[(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"namespace"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u76d1\u542c\u7684\u547d\u540d\u7a7a\u95f4\uff0c\u7559\u7a7a\u8868\u793a\u76d1\u542c\u6240\u6709\u547d\u540d\u7a7a\u95f4"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u7a7a"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u7a7a"})]}),(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"manager-namespace"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"ob-operator \u8fd0\u884c\u7684\u547d\u540d\u7a7a\u95f4"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"oceanbase-system"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"oceanbase-system"})]}),(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"metrics-bind-address"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"ob-operator \u63d0\u4f9b Prometheus \u6307\u6807\u7684\u670d\u52a1\u7aef\u53e3"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:":8080"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"127.0.0.1:8080"})]}),(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"health-probe-bind-address"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"ob-operator \u8fdb\u7a0b\u5065\u5eb7\u63a2\u9488\u7ed1\u5b9a\u7aef\u53e3"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:":8081"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:":8081"})]}),(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"leader-elect"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u542f\u52a8 ob-operator \u65f6\u662f\u5426\u91c7\u7528\u9009\u4e3b\u6d41\u7a0b"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"false"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"true"})]}),(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"log-verbosity"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u65e5\u5fd7\u8f93\u51fa\u91cf\uff0c\u4e3a 0 \u8f93\u51fa\u5173\u952e\u4fe1\u606f\uff0c\u4e3a 1 \u8f93\u51fa\u8c03\u8bd5\u4fe1\u606f\uff0c\u4e3a 2 \u8f93\u51fa\u6eaf\u6e90\u4fe1\u606f"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"0"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"0"})]})]})]}),"\n",(0,l.jsx)(t.h2,{id:"\u73af\u5883\u53d8\u91cf",children:"\u73af\u5883\u53d8\u91cf"}),"\n",(0,l.jsxs)(t.table,{children:[(0,l.jsx)(t.thead,{children:(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.th,{style:{textAlign:"left"},children:"\u73af\u5883\u53d8\u91cf\u540d"}),(0,l.jsx)(t.th,{style:{textAlign:"left"},children:"\u542b\u4e49"}),(0,l.jsx)(t.th,{style:{textAlign:"left"},children:"deploy \u914d\u7f6e"})]})}),(0,l.jsxs)(t.tbody,{children:[(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"TELEMETRY_REPORT_HOST"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u9065\u6d4b\u91c7\u96c6\u6570\u636e\u6536\u96c6\u7aef"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:(0,l.jsx)(t.a,{href:"https://openwebapi.oceanbase.com",children:"https://openwebapi.oceanbase.com"})})]}),(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"TELEMETRY_DEBUG"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u8bbe\u7f6e\u4e3a true \u53ef\u5f00\u542f\u9065\u6d4b\u91c7\u96c6\u8c03\u8bd5\u6a21\u5f0f"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u7a7a"})]}),(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"DISABLE_WEBHOOKS"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u8bbe\u7f6e\u4e3a true \u53ef\u7981\u7528 webhooks \u6821\u9a8c"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u7a7a"})]}),(0,l.jsxs)(t.tr,{children:[(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"DISABLE_TELEMETRY"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u8bbe\u7f6e\u4e3a true \u53ef\u7981\u7528\u9065\u6d4b\u91c7\u96c6\u6a21\u5757\uff0c\u9065\u6d4b\u91c7\u96c6\u6a21\u5757\u4f1a\u91c7\u96c6\u96c6\u7fa4\u73af\u5883\u548c\u4e8b\u4ef6\u4fe1\u606f\u8131\u654f\u540e\u53d1\u9001\u7ed9 OceanBase\uff0c\u671f\u671b\u901a\u8fc7\u8fd9\u4e9b\u6570\u636e\u5e2e\u52a9\u6539\u5584 ob-operator"}),(0,l.jsx)(t.td,{style:{textAlign:"left"},children:"\u7a7a"})]})]})]}),"\n",(0,l.jsx)(t.h2,{id:"\u4fee\u6539\u65b9\u6cd5",children:"\u4fee\u6539\u65b9\u6cd5"}),"\n",(0,l.jsxs)(t.p,{children:["\u4f7f\u7528 ",(0,l.jsx)(t.code,{children:"deploy/operator.yaml"})," \u4e2d\u7684\u914d\u7f6e\u6587\u4ef6\uff0c\u627e\u5230\u540d\u4e3a ",(0,l.jsx)(t.code,{children:"oceanbase-controller-manager"})," \u7684 ",(0,l.jsx)(t.code,{children:"Deployment"})," \u8d44\u6e90\uff0c\u5728\u5176\u5bb9\u5668\u5217\u8868\u4e2d\u4fee\u6539\u540d\u4e3a ",(0,l.jsx)(t.code,{children:"manager"})," \u5bb9\u5668\u7684\u542f\u52a8\u53c2\u6570\u548c\u73af\u5883\u53d8\u91cf\uff0c\u4e0b\u9762\u622a\u53d6 ",(0,l.jsx)(t.code,{children:"deploy/operator.yaml"})," \u4e2d\u8be5\u90e8\u5206\u4e3a\u4f8b\u3002"]}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-yaml",children:"      # \u539f\u672c\u7684\u914d\u7f6e\n      containers:\n      - args:\n        - --health-probe-bind-address=:8081\n        - --metrics-bind-address=:8080\n        - --leader-elect\n        - --manager-namespace=oceanbase-system\n        - --log-verbosity=0\n        command:\n        - /manager\n        env:\n        - name: TELEMETRY_REPORT_HOST\n          value: https://openwebapi.oceanbase.com\n"})}),"\n",(0,l.jsx)(t.h3,{id:"\u793a\u4f8b\u589e\u5927\u65e5\u5fd7\u8f93\u51fa\u91cf",children:"\u793a\u4f8b\uff1a\u589e\u5927\u65e5\u5fd7\u8f93\u51fa\u91cf"}),"\n",(0,l.jsxs)(t.p,{children:["\u5982\u679c\u7528\u6237\u5e0c\u671b\u589e\u52a0 ob-operator \u7684\u65e5\u5fd7\u8f93\u51fa\u91cf\uff0c\u53ef\u589e\u5927 ",(0,l.jsx)(t.code,{children:"log-verbosity"})," \u53c2\u6570\u5230 1 \u6216\u8005 2\uff0c\u503c\u8d8a\u5927\u8f93\u51fa\u7684\u65e5\u5fd7\u8d8a\u591a\u3002"]}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-yaml",children:"      # \u4fee\u6539\u540e\u7684\u914d\u7f6e\n      containers:\n      - args:\n        - --health-probe-bind-address=:8081\n        - --metrics-bind-address=:8080\n        - --leader-elect\n        - --manager-namespace=oceanbase-system\n        - --log-verbosity=2 # \u65e5\u5fd7\u8f93\u51fa\u91cf\u589e\u5927\u5230 2 \n        command:\n        - /manager\n        env:\n        - name: TELEMETRY_REPORT_HOST\n          value: https://openwebapi.oceanbase.com\n"})}),"\n",(0,l.jsx)(t.h3,{id:"\u793a\u4f8b\u6307\u5b9a\u8d44\u6e90\u547d\u540d\u7a7a\u95f4",children:"\u793a\u4f8b\uff1a\u6307\u5b9a\u8d44\u6e90\u547d\u540d\u7a7a\u95f4"}),"\n",(0,l.jsx)(t.pre,{children:(0,l.jsx)(t.code,{className:"language-yaml",children:"      # \u4fee\u6539\u540e\u7684\u914d\u7f6e\n      containers:\n      - args:\n        - --health-probe-bind-address=:8081\n        - --metrics-bind-address=:8080\n        - --leader-elect\n        - --manager-namespace=oceanbase-system\n        - --log-verbosity=0\n        - --namespace=oceanbase # \u9650\u5b9a ob-operator \u53ea\u76d1\u542c\u547d\u540d\u7a7a\u95f4\u4e3a oceanbase \u5185\u7684\u8d44\u6e90\n        command:\n        - /manager\n        env:\n        - name: TELEMETRY_REPORT_HOST\n          value: https://openwebapi.oceanbase.com\n"})}),"\n",(0,l.jsx)(t.h3,{id:"\u5e94\u7528\u5230\u96c6\u7fa4\u4e2d",children:"\u5e94\u7528\u5230\u96c6\u7fa4\u4e2d"}),"\n",(0,l.jsxs)(t.p,{children:["\u4fee\u6539\u5b8c\u6210\u540e\u901a\u8fc7 ",(0,l.jsx)(t.code,{children:"kubectl apply -f deploy/operator.yaml"})," \u5c06\u914d\u7f6e\u6587\u4ef6\u5e94\u7528\u5230\u96c6\u7fa4\u4e2d\u5373\u53ef\u751f\u6548\u3002\u73af\u5883\u53d8\u91cf\u7684\u914d\u7f6e\u65b9\u6cd5\u4e0e\u542f\u52a8\u53c2\u6570\u76f8\u540c\uff0c\u672c\u6587\u4e0d\u518d\u8d58\u8ff0\u3002"]})]})}function h(e={}){const{wrapper:t}={...(0,s.R)(),...e.components};return t?(0,l.jsx)(t,{...e,children:(0,l.jsx)(c,{...e})}):c(e)}},8453:(e,t,n)=>{n.d(t,{R:()=>a,x:()=>o});var l=n(6540);const s={},r=l.createContext(s);function a(e){const t=l.useContext(r);return l.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function o(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:a(e.components),l.createElement(r.Provider,{value:t},e.children)}}}]);
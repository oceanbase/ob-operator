"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[8524],{2937:(e,n,o)=>{o.r(n),o.d(n,{assets:()=>l,contentTitle:()=>i,default:()=>u,frontMatter:()=>t,metadata:()=>s,toc:()=>c});var r=o(4848),a=o(8453);const t={},i="Develop ob-operator locally",s={id:"developer/develop-locally",title:"Develop ob-operator locally",description:"Introduction",source:"@site/i18n/zh-Hans/docusaurus-plugin-content-docs/current/developer/develop-locally.md",sourceDirName:"developer",slug:"/developer/develop-locally",permalink:"/ob-operator/zh-Hans/docs/developer/develop-locally",draft:!1,unlisted:!1,editUrl:"https://github.com/oceanbase/ob-operator/tree/master/docsite/docs/developer/develop-locally.md",tags:[],version:"current",frontMatter:{},sidebar:"developerSidebar",previous:{title:"ob-operator \u90e8\u7f72",permalink:"/ob-operator/zh-Hans/docs/developer/deploy"}},l={},c=[{value:"Introduction",id:"introduction",level:2},{value:"Requirements",id:"requirements",level:2},{value:"Modify and Apply CRDs",id:"modify-and-apply-crds",level:2},{value:"Modify definition of CRDs",id:"modify-definition-of-crds",level:3},{value:"Apply the Changes",id:"apply-the-changes",level:3},{value:"Modify and Run ob-operator",id:"modify-and-run-ob-operator",level:2},{value:"Build and Deploy",id:"build-and-deploy",level:3},{value:"Run locally",id:"run-locally",level:3},{value:"Disable Webhook and CertManager",id:"disable-webhook-and-certmanager",level:4},{value:"Generate Self-signed Certificates",id:"generate-self-signed-certificates",level:4},{value:"Run ob-operator on Your Machine",id:"run-ob-operator-on-your-machine",level:4},{value:"Debugging",id:"debugging",level:3},{value:"Start delve Debug Server",id:"start-delve-debug-server",level:4},{value:"Debug in Terminal",id:"debug-in-terminal",level:4},{value:"Debug in VSCode",id:"debug-in-vscode",level:4}];function d(e){const n={a:"a",admonition:"admonition",code:"code",h1:"h1",h2:"h2",h3:"h3",h4:"h4",img:"img",li:"li",p:"p",pre:"pre",ul:"ul",...(0,a.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(n.h1,{id:"develop-ob-operator-locally",children:"Develop ob-operator locally"}),"\n",(0,r.jsx)(n.h2,{id:"introduction",children:"Introduction"}),"\n",(0,r.jsxs)(n.p,{children:["ob-operator depends on ",(0,r.jsx)(n.a,{href:"https://kubebuilder.io/introduction",children:"kubebuilder"}),", an operator framework maintained by kubernetes SIGS. It offers convenient utilities to bootstrap an operator and manage API types in it. Like other operator frameworks, kubebuilder depends on kubernetes ",(0,r.jsx)(n.a,{href:"https://github.com/kubernetes-sigs/controller-runtime",children:"controller runtime"})," either, which is an excellent reference to know how kubernetes dispatch events and reconcile resources."]}),"\n",(0,r.jsx)(n.h2,{id:"requirements",children:"Requirements"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:["Go 1.22 or above is required to build ob-operator, you can refer to the ",(0,r.jsx)(n.a,{href:"https://go.dev/",children:"official website"})," to setup go environment."]}),"\n",(0,r.jsxs)(n.li,{children:[(0,r.jsx)(n.a,{href:"https://www.gnu.org/software/make/",children:"make"})," is used for a variety of build and test actions."]}),"\n",(0,r.jsxs)(n.li,{children:["kubebuilder is used as k8s operator framework. it's highly recommended to read ",(0,r.jsx)(n.a,{href:"https://book.kubebuilder.io/",children:"kubebuilder books"}),"."]}),"\n",(0,r.jsxs)(n.li,{children:["Access to a kubernetes cluster is required to develop and test the operator. You can use ",(0,r.jsx)(n.a,{href:"https://minikube.sigs.k8s.io/docs/",children:"minikube"}),", ",(0,r.jsx)(n.a,{href:"https://kind.sigs.k8s.io/docs/user/quick-start/",children:"kind"})," or ",(0,r.jsx)(n.a,{href:"https://k3s.io/",children:"k3s"})," to create a local kubernetes cluster."]}),"\n"]}),"\n",(0,r.jsx)(n.h2,{id:"modify-and-apply-crds",children:"Modify and Apply CRDs"}),"\n",(0,r.jsx)(n.h3,{id:"modify-definition-of-crds",children:"Modify definition of CRDs"}),"\n",(0,r.jsxs)(n.p,{children:[(0,r.jsx)(n.code,{children:"ob-operator"})," uses kubebuilder as operator framework, which generates CRDs in YAML format from Go code lies in ",(0,r.jsx)(n.code,{children:"api/v1alpha1"})," directory. If you have modified the Go code, for example adding a new field to ",(0,r.jsx)(n.code,{children:"OBClusterSpec"})," in ",(0,r.jsx)(n.code,{children:"api/v1alpha1/obcluster_types.go"}),", you need to regenerate the CRDs by running ",(0,r.jsx)(n.code,{children:"make generate manifests fmt"}),". Then you will see changes in ",(0,r.jsx)(n.code,{children:"config/crd/bases/oceanbase.oceanbase.com_xxx.yaml"})," files."]}),"\n",(0,r.jsx)(n.h3,{id:"apply-the-changes",children:"Apply the Changes"}),"\n",(0,r.jsxs)(n.p,{children:["You can apply the changes of CRDs to kubernetes cluster by running ",(0,r.jsx)(n.code,{children:"make install"}),". This command will apply the CRDs to the kubernetes cluster."]}),"\n",(0,r.jsxs)(n.p,{children:["You will see output like the following after executing ",(0,r.jsx)(n.code,{children:"kubectl get crds"})," command if the CRDs are successfully applied:"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"$ kubectl get crds | grep oceanbase.oceanbase.com\nobclusters.oceanbase.oceanbase.com               2024-01-01T00:00:00Z\nobparameters.oceanbase.oceanbase.com             2024-01-01T00:00:00Z\nobresourcerescues.oceanbase.oceanbase.com        2024-01-01T00:00:00Z\nobservers.oceanbase.oceanbase.com                2024-01-01T00:00:00Z\nobtenantbackuppolicies.oceanbase.oceanbase.com   2024-01-01T00:00:00Z\nobtenantbackups.oceanbase.oceanbase.com          2024-01-01T00:00:00Z\nobtenantoperations.oceanbase.oceanbase.com       2024-01-01T00:00:00Z\nobclusteroperations.oceanbase.oceanbase.com      2024-01-01T00:00:00Z\nobtenantrestores.oceanbase.oceanbase.com         2024-01-01T00:00:00Z\nobtenants.oceanbase.oceanbase.com                2024-01-01T00:00:00Z\nobzones.oceanbase.oceanbase.com                  2024-01-01T00:00:00Z\n"})}),"\n",(0,r.jsx)(n.h2,{id:"modify-and-run-ob-operator",children:"Modify and Run ob-operator"}),"\n",(0,r.jsxs)(n.p,{children:["ob-operator acts as a controller manager, which watches the resources in kubernetes cluster and reconciles them. The controller manager is generated by kubebuilder, and the main logic lies in ",(0,r.jsx)(n.code,{children:"internal/{controller,resource}"})," and ",(0,r.jsx)(n.code,{children:"pkg/coordinator"})," directories."]}),"\n",(0,r.jsx)(n.h3,{id:"build-and-deploy",children:"Build and Deploy"}),"\n",(0,r.jsxs)(n.p,{children:["After modifying the code, you can build the docker image by running ",(0,r.jsx)(n.code,{children:"make docker-build docker-push IMG=<your-image-name>"}),". Then you can deploy the controller manager to the kubernetes cluster by running ",(0,r.jsx)(n.code,{children:"make deploy IMG=<your-image-name>"}),"."]}),"\n",(0,r.jsxs)(n.p,{children:["If the docker need root permission to build, you can run building command with ",(0,r.jsx)(n.code,{children:"sudo"}),", like"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-bash",children:"sudo make docker-build IMG=<your-image-name>\nmake docker-push IMG=<your-image-name> # If you need to push the image to a registry\nmake deploy IMG=<your-image-name>\n"})}),"\n",(0,r.jsx)(n.admonition,{type:"tip",children:(0,r.jsxs)(n.p,{children:["If the developing machine and the deploying machine is the same one, you can skip the pushing step ",(0,r.jsx)(n.code,{children:"make docker-push"}),"."]})}),"\n",(0,r.jsx)(n.h3,{id:"run-locally",children:"Run locally"}),"\n",(0,r.jsx)(n.p,{children:"Building docker image and pushing it to a registry is time-consuming, especially when you are developing and debugging. You can run controller manager locally to accelerate the development process."}),"\n",(0,r.jsx)(n.admonition,{type:"tip",children:(0,r.jsx)(n.p,{children:"In this step, we'll disable webhook validation and run controller manager on the developing machine. The controller manager will communicate with kubernetes cluster by local .kube/config configuration file."})}),"\n",(0,r.jsx)(n.h4,{id:"disable-webhook-and-certmanager",children:"Disable Webhook and CertManager"}),"\n",(0,r.jsxs)(n.p,{children:["There are many configuration items that marked by ",(0,r.jsx)(n.code,{children:"[CERTMANAGER]"})," and ",(0,r.jsx)(n.code,{children:"[WEBHOOK]"})," in the two files ",(0,r.jsx)(n.code,{children:"config/crd/kustomization.yaml"})," and ",(0,r.jsx)(n.code,{children:"config/default/kustomization.yaml"}),". They are used to enable and configure webhooks in real kubernetes deployment. Because we want to run controller manager locally, we need to disable them."]}),"\n",(0,r.jsxs)(n.p,{children:["You could just apply the latest ",(0,r.jsx)(n.code,{children:"deploy/operator.yaml"})," manifest and delete the following resources to deploy CRDs and make controller manager uninstalled."]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"kubectl delete validatingwebhookconfigurations.admissionregistration.k8s.io oceanbase-validating-webhook-configuration\nkubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io oceanbase-mutating-webhook-configuration\nkubectl delete -n oceanbase-system svc oceanbase-webhook-service\nkubectl delete -n oceanbase-system deployments.apps oceanbase-controller-manager\n"})}),"\n",(0,r.jsx)(n.h4,{id:"generate-self-signed-certificates",children:"Generate Self-signed Certificates"}),"\n",(0,r.jsx)(n.p,{children:"It's necessary for node hosting controller manager to have a TLS certificate. In the real kubernetes cluster, the cert-manager will inject the sign into the controller manager pod. On our laptop, we need self-sign one:"}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"mkdir -p /tmp/k8s-webhook-server/serving-certs\nopenssl genrsa -out /tmp/k8s-webhook-server/serving-certs/tls.key 2048\nopenssl req -new -key /tmp/k8s-webhook-server/serving-certs/tls.key -out /tmp/k8s-webhook-server/serving-certs/tls.csr\nopenssl x509 -req -days 365 -in /tmp/k8s-webhook-server/serving-certs/tls.csr -signkey /tmp/k8s-webhook-server/serving-certs/tls.key -out /tmp/k8s-webhook-server/serving-certs/tls.crt\n"})}),"\n",(0,r.jsx)(n.h4,{id:"run-ob-operator-on-your-machine",children:"Run ob-operator on Your Machine"}),"\n",(0,r.jsxs)(n.p,{children:["There are some useful commands in ",(0,r.jsx)(n.code,{children:"Makefile"})," and ",(0,r.jsx)(n.code,{children:"make/*.mk"}),", we could type ",(0,r.jsx)(n.code,{children:"make run-local"})," to start controller manager locally. Or redirect the output to a log file for better analytic experience,"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"# print log to stdout\nmake run-local\n# or redirect output to a file\nmake run-local &> bin/run.log\n"})}),"\n",(0,r.jsx)(n.h3,{id:"debugging",children:"Debugging"}),"\n",(0,r.jsxs)(n.p,{children:["Though print debugging is enough for most cases, there are quite some cases that are not obvious from printed information. We could debug with go debugging tool ",(0,r.jsx)(n.a,{href:"https://github.com/go-delve/delve",children:"delve"}),"."]}),"\n",(0,r.jsxs)(n.p,{children:[(0,r.jsx)(n.code,{children:"install-delve"})," command is declared in ",(0,r.jsx)(n.code,{children:"make/debug.mk"}),", you can type ",(0,r.jsx)(n.code,{children:"make install-delve"})," to get it. The help message of it can be glanced,"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",metastring:"dlv help",children:"Delve is a source level debugger for Go programs.\n\nDelve enables you to interact with your program by controlling the execution of the process,\nevaluating variables, and providing information of thread / goroutine state, CPU register state and more.\n\nThe goal of this tool is to provide a simple yet powerful interface for debugging Go programs.\n\nPass flags to the program you are debugging using `--`, for example:\n\n`dlv exec ./hello -- server --config conf/config.toml`\n\nUsage:\n  dlv [command]\n\nAvailable Commands:\n  attach      Attach to running process and begin debugging.\n  completion  Generate the autocompletion script for the specified shell\n  connect     Connect to a headless debug server with a terminal client.\n  core        Examine a core dump.\n  dap         Starts a headless TCP server communicating via Debug Adaptor Protocol (DAP).\n  debug       Compile and begin debugging main package in current directory, or the package specified.\n  exec        Execute a precompiled binary, and begin a debug session.\n  help        Help about any command\n  test        Compile test binary and begin debugging program.\n  trace       Compile and begin tracing program.\n  version     Prints version.\n\nAdditional help topics:\n  dlv backend    Help about the --backend flag.\n  dlv log        Help about logging flags.\n  dlv redirect   Help about file redirection.\n"})}),"\n",(0,r.jsx)(n.h4,{id:"start-delve-debug-server",children:"Start delve Debug Server"}),"\n",(0,r.jsxs)(n.p,{children:["Run ",(0,r.jsx)(n.code,{children:"make run-delve"})," simply to start debugging server."]}),"\n",(0,r.jsx)(n.h4,{id:"debug-in-terminal",children:"Debug in Terminal"}),"\n",(0,r.jsxs)(n.p,{children:["If you prefer to debug in terminal, with ",(0,r.jsx)(n.code,{children:"dlv connect 127.0.0.1:2345"})," command you can connect to the debugging server. After connecting, you enter a REPL environment of delve, available commands are showed below,"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"(dlv) help\nThe following commands are available:\n\nRunning the program:\n    call ------------------------ Resumes process, injecting a function call (EXPERIMENTAL!!!)\n    continue (alias: c) --------- Run until breakpoint or program termination.\n    next (alias: n) ------------- Step over to next source line.\n    rebuild --------------------- Rebuild the target executable and restarts it. It does not work if the executable was not built by delve.\n    restart (alias: r) ---------- Restart process.\n    step (alias: s) ------------- Single step through program.\n    step-instruction (alias: si)  Single step a single cpu instruction.\n    stepout (alias: so) --------- Step out of the current function.\n\nManipulating breakpoints:\n    break (alias: b) ------- Sets a breakpoint.\n    breakpoints (alias: bp)  Print out info for active breakpoints.\n    clear ------------------ Deletes breakpoint.\n    clearall --------------- Deletes multiple breakpoints.\n    condition (alias: cond)  Set breakpoint condition.\n    on --------------------- Executes a command when a breakpoint is hit.\n    toggle ----------------- Toggles on or off a breakpoint.\n    trace (alias: t) ------- Set tracepoint.\n    watch ------------------ Set watchpoint.\n\nViewing program variables and memory:\n    args ----------------- Print function arguments.\n    display -------------- Print value of an expression every time the program stops.\n    examinemem (alias: x)  Examine raw memory at the given address.\n    locals --------------- Print local variables.\n    print (alias: p) ----- Evaluate an expression.\n    regs ----------------- Print contents of CPU registers.\n    set ------------------ Changes the value of a variable.\n    vars ----------------- Print package variables.\n    whatis --------------- Prints type of an expression.\n\nListing and switching between threads and goroutines:\n    goroutine (alias: gr) -- Shows or changes current goroutine\n    goroutines (alias: grs)  List program goroutines.\n    thread (alias: tr) ----- Switch to the specified thread.\n    threads ---------------- Print out info for every traced thread.\n\nViewing the call stack and selecting frames:\n    deferred --------- Executes command in the context of a deferred call.\n    down ------------- Move the current frame down.\n    frame ------------ Set the current frame, or execute command on a different frame.\n    stack (alias: bt)  Print stack trace.\n    up --------------- Move the current frame up.\n\nOther commands:\n    config --------------------- Changes configuration parameters.\n    disassemble (alias: disass)  Disassembler.\n    dump ----------------------- Creates a core dump from the current process state\n    edit (alias: ed) ----------- Open where you are in $DELVE_EDITOR or $EDITOR\n    exit (alias: quit | q) ----- Exit the debugger.\n    funcs ---------------------- Print list of functions.\n    help (alias: h) ------------ Prints the help message.\n    libraries ------------------ List loaded dynamic libraries\n    list (alias: ls | l) ------- Show source code.\n    packages ------------------- Print list of packages.\n    source --------------------- Executes a file containing a list of delve commands\n    sources -------------------- Print list of source files.\n    target --------------------- Manages child process debugging.\n    transcript ----------------- Appends command output to a file.\n    types ---------------------- Print list of types\n\nType help followed by a command for full documentation.\n"})}),"\n",(0,r.jsx)(n.h4,{id:"debug-in-vscode",children:"Debug in VSCode"}),"\n",(0,r.jsxs)(n.p,{children:["If you are first to debugging in VSCode, enter ",(0,r.jsx)(n.code,{children:"Cmd+Shift+P"})," to open commands panel. Then, type ",(0,r.jsx)(n.code,{children:"Debug: Add Configuration..."})," and create debugging task for Go. After creating task successfully, open commands panel and type ",(0,r.jsx)(n.code,{children:"Debug: Start Debugging"}),"/",(0,r.jsx)(n.code,{children:"Debug: Select and Start Debugging"})," to start debugging."]}),"\n",(0,r.jsx)(n.p,{children:(0,r.jsx)(n.img,{alt:"Debug in VSCode",src:o(4361).A+"",width:"1500",height:"988"})})]})}function u(e={}){const{wrapper:n}={...(0,a.R)(),...e.components};return n?(0,r.jsx)(n,{...e,children:(0,r.jsx)(d,{...e})}):d(e)}},4361:(e,n,o)=>{o.d(n,{A:()=>r});const r=o.p+"assets/images/debug-in-vscode-e840b06a074b2381fab93973768f71ac.png"},8453:(e,n,o)=>{o.d(n,{R:()=>i,x:()=>s});var r=o(6540);const a={},t=r.createContext(a);function i(e){const n=r.useContext(t);return r.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function s(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:i(e.components),r.createElement(t.Provider,{value:n},e.children)}}}]);
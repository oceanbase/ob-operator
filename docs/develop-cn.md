# 如何参与开发

## 代码目录结构

```yaml
.
├── Dockerfile
├── LEGAL.md
├── LICENSE
├── LICENSE.Apache
├── LICENSE.MIT
├── Makefile
├── PROJECT
├── README-CN.md
├── README.md
├── apis  // CRD 定义
├── cmd
├── config  // kustomize 相关配置与生成文件
├── deploy  // 部署服务所使用的的文件
├── docs  // 文档
├── go.mod
├── go.sum
├── hack
├── main.go
├── pkg
│ ├── cable  // 守护进程逻辑
│ ├── config
│ ├── controllers
│ │ ├── observer  // observer controller 逻辑
│ │ │ └── cable  // 与 Agent 进程交互的部分
│ │ └── statefulapp  // statefulapp controller 逻辑
│ ├── infrastructure
│ │ ├── kube
│ │ └── ob
│ ├── kubeclient
│ └── util
├── scripts
│ └── observer
└── test
    └── e2e
```

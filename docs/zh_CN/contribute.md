# 如何参与开发

[English version](../en_US/contribute.md)

## 代码目录结构

```yaml
.
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
├── PROJECT
├── api
│   ├── constants
│   ├── types
│   └── v1alpha1
├── charts
│   ├── ob-operator
│   └── oceanbase-cluster
├── cmd
│   └── main.go
├── config/
├── deploy/
├── distribution
│   ├── obagent
│   ├── ob-configserver
│   ├── obproxy
│   └── oceanbase
├── doc/
├── example/
├── hack/
├── make/
├── pkg
│   ├── const
│   ├── controller
│   ├── database
│   ├── oceanbase
│   ├── resource
│   ├── task
│   └── ...
└── README.md
```

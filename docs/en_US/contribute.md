# How to contribute

该文档也提供[中文版](../zh_CN/contribute.md)。

## Code structure

```yaml
.
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
├── PROJECT
├── api # CRD definition and Webhook implementation
│   ├── constants
│   ├── types
│   └── v1alpha1
├── charts # Helm chart implementation
│   ├── ob-operator
│   └── oceanbase-cluster
├── cmd # Command directory, program entry
│   └── main.go
├── config/ # kustomize configuration directory
├── deploy/ # YAML files required for ob-operator deployment
├── distribution # Kubernetes-related component build files
├── doc/ # Documentation directory
├── example/ # Example configuration files
├── hack/ 
├── make/
├── pkg
│   ├── const
│   ├── controller # Controller implementations
│   ├── database # Database connection pool
│   ├── oceanbase # OceanBase SDK
│   ├── resource # Resource manager interfaces and implementations
│   ├── task # Task flow and task manager
│   └── ...
└── README.md
```

## Contribution

- [Submit an issue](https://github.com/oceanbase/ob-operator/issues)
- [Create a Pull request](https://github.com/oceanbase/ob-operator/pulls)
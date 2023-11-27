# 如何参与开发

[English version](../en_US/contribute.md) is available.

## 代码目录结构

```yaml
.
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
├── PROJECT
├── api # CRD 定义和 Webhook 实现
│   ├── constants
│   ├── types
│   └── v1alpha1
├── charts # Helm chart 实现
│   ├── ob-operator
│   └── oceanbase-cluster
├── cmd # 命令目录，程序入口
│   └── main.go
├── config/ # kustomize 配置目录
├── deploy/ # 部署 ob-operator 所需的 YAML 文件
├── distribution # Kubernetes 相关的组件构建文件
├── doc/ # 文档目录
├── example/ # 示例配置文件
├── hack/ 
├── make/
├── pkg
│   ├── const
│   ├── controller # 控制器的实现
│   ├── database # 数据库连接池
│   ├── oceanbase # OceanBase SDK
│   ├── resource # 资源管理器接口和各个资源管理器的实现
│   ├── task # 任务流和任务管理器
│   └── ...
└── README.md
```

## 参与方式

- [提出 Issue](https://github.com/oceanbase/ob-operator/issues)
- [发起 Pull request](https://github.com/oceanbase/ob-operator/pulls)
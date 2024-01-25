# GSoC 2024 Ideas List

Hi there! This is the ideas list of OceanBase for Google Summer of Code 2024! 

As a first-year mentor organization, we focus the development ideas on ob-operator, a kubernetes operator which helps deploy and manage OceanBase cluster on kubernetes cluster seamlessly. There are four projects for contributors, and we'll offer delicate guidance for every choice.

Enjoy your summer of code!

## 1. CLI tool

* Project Description: The primary method to manage CRDs of ob-operator is manipulating YAML manifests which is not easy enough for new-coming users to get started. Within the CLI tool, features like component installation, demo setup,  resource management and necessary validation can be implemented.
* Required Skills: Golang, Kubernetes
* Project Size: medium/large
* Expected Outcomes: A complete CLI tool to control and manage CRDs in ob-operator, in other words, to manage clusters, tenants, backups on ob-operator in kubernetes cluster. Or a module in the CLI tool.
* References: 
  * [kubernetes/client-go](https://github.com/kubernetes/client-go)
  * [spf13/cobra](https://github.com/spf13/cobra)
  * [manifoldco/promptui](https://github.com/manifoldco/promptui)


## 2. Light-weighted Operations

* Project Description: Use Light-weighted operation task types to influence the status of resources instead of modifying specifications of those bigger resources.
* Required Skills: Golang, Development of Kubernetes controller, Docker
* Project Size: medium
* Expected Outcomes: Implementation of CRD(s) and corresponding controller(s) for cluster and tenant, which can trigger small operational actions and reveal the progress through `status` fields.
* References: 
  * [Tenant operation: Failover](https://en.oceanbase.com/docs/common-oceanbase-database-10000000001106036)
  * [Tenant operation: Replay log](https://en.oceanbase.com/docs/common-oceanbase-database-10000000001103949)
  * [Architecture of ob-operator](https://oceanbase.github.io/ob-operator/docs/en_US/arch.html)


## 3. Alertmanager Integration

* Project Description: There are some features about cluster management in OceanBase dashboard, a web-based dashboard application, and tenant management features are coming soon. OceanBase dashboard has bundled prometheus to manage time series metrics data by now. It's recommended that Alertmanager performs alerting tasks along with promethues. Integrating Alertmanager into OceanBase dashboard could be an attracive feature.
* Required Skills: Web development(React), Kubernetes, Prometheus, Alertmanager
* Project Size: medium
* Expected Outcomes: A functional panel in OceanBase dashboard in which users could view alert events, configure alert rules, and define alert templates in the scope of Alertmanager.
* References: 
  * [Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager/)
  * [prometheus/alertmanager](https://github.com/prometheus/alertmanager)
  * [OceanBase - Monitor - Overview](https://en.oceanbase.com/docs/common-oceanbase-database-10000000001103563)


## 4. OceanBase Database Proxy integration

* Project Description: OceanBase Database Proxy (ODP), also called OBProxy, is a dedicated proxy server for OceanBase Database. Core features of ODP include connection management, optimal routing, high-performance forwarding, easy O&M, high availability, and proprietary protocol. ODP should be integrated into OceanBase dashboard to enhance proxy management.
* Required Skills: Web development(React), Golang Kubernetes,
* Project Size: medium
* Expected Outcomes: A functional panel in OceanBase dashboard where users could setup, configure and even delete ODP for specific OceanBase cluster.
* References: 
  * [OceanBase Database Proxy](https://en.oceanbase.com/docs/odp-en)


## 5. Accounts management and RBAC admission control

* Project Description: OceanBase Dashboard is a web-based management application for ob-operator, with support for managing cluster and tenant clearly. Currently, it has a quite simple account system that stores and retrieves user credentials with `Secret` resource. And, it lacks any form of admission control. So it would be a good starting point to develop an account management module, complemented by a robust RBAC (Role-Based Access Control) permissions system.
* Required Skills: Golang, Kubernetes
* Project Size: medium
* Expected Outcomes: an advanced account management module paired with an RBAC-aligned permissions system.
* References: 
  * [casbin/casbin](https://github.com/casbin/casbin)
  * [Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
  * [Certificates and Certificate Signing Requests](https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/)
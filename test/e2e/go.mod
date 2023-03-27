module github.com/oceanbase/ob-operator/test/e2e

go 1.16

require (
	github.com/oceanbase/ob-operator v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/prometheus/client_golang v1.11.0 // indirect
	go.uber.org/zap v1.19.0 // indirect
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)

replace github.com/oceanbase/ob-operator => ../../

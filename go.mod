module github.com/oceanbase/ob-operator

go 1.22

require (
	github.com/deckarep/golang-set v1.8.0
	github.com/gin-contrib/gzip v0.0.6
	github.com/gin-contrib/requestid v0.0.6
	github.com/gin-contrib/sessions v0.0.5
	github.com/gin-contrib/static v1.1.0
	github.com/gin-gonic/gin v1.9.1
	github.com/go-logr/logr v1.2.4
	github.com/go-resty/resty/v2 v2.11.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/google/uuid v1.4.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/jmoiron/sqlx v1.3.5
	github.com/mattn/go-isatty v0.0.20
	github.com/mitchellh/mapstructure v1.5.0
	github.com/onsi/ginkgo/v2 v2.11.0
	github.com/onsi/gomega v1.27.10
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.18.2
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.3
	go.uber.org/zap v1.24.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.27.2
	k8s.io/apimachinery v0.27.2
	k8s.io/client-go v0.27.2
	k8s.io/kubernetes v1.27.2
	sigs.k8s.io/controller-runtime v0.15.0
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.10.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/zapr v1.2.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.18.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/prometheus/client_golang v1.15.1 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/arch v0.7.0 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/oauth2 v0.15.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/term v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.27.2 // indirect
	k8s.io/component-base v0.27.2 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/utils v0.0.0-20230209194617-a36077c30491 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace k8s.io/api => k8s.io/api v0.27.2

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.27.2

replace k8s.io/apimachinery => k8s.io/apimachinery v0.27.2

replace k8s.io/apiserver => k8s.io/apiserver v0.27.2

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.27.2

replace k8s.io/client-go => k8s.io/client-go v0.27.2

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.27.2

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.27.2

replace k8s.io/code-generator => k8s.io/code-generator v0.27.2

replace k8s.io/component-base => k8s.io/component-base v0.27.2

replace k8s.io/component-helpers => k8s.io/component-helpers v0.27.2

replace k8s.io/controller-manager => k8s.io/controller-manager v0.27.2

replace k8s.io/cri-api => k8s.io/cri-api v0.28.0-alpha.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.27.2

replace k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.27.2

replace k8s.io/kms => k8s.io/kms v0.27.2

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.27.2

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.27.2

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.27.2

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.27.2

replace k8s.io/kubectl => k8s.io/kubectl v0.27.2

replace k8s.io/kubelet => k8s.io/kubelet v0.27.2

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.27.2

replace k8s.io/metrics => k8s.io/metrics v0.27.2

replace k8s.io/mount-utils => k8s.io/mount-utils v0.27.2

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.27.2

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.27.2

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.27.2

replace k8s.io/sample-controller => k8s.io/sample-controller v0.27.2

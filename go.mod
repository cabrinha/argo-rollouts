module github.com/argoproj/argo-rollouts

go 1.13

require (
	github.com/antonmedv/expr v1.4.2
	github.com/bombsimon/wsl/v2 v2.2.0 // indirect
	github.com/bouk/monkey v1.0.0
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20190327010347-be7ac8be2ae0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/emicklei/go-restful v2.11.2+incompatible // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-openapi/spec v0.19.6
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/gobuffalo/flect v0.2.1 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/mock v1.4.1 // indirect
	github.com/golangci/gocyclo v0.0.0-20180528144436-0a533e8fa43d // indirect
	github.com/golangci/golangci-lint v1.23.8 // indirect
	github.com/golangci/revgrep v0.0.0-20180812185044-276a5c0a1039 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/jirfag/go-printf-func-name v0.0.0-20200119135958-7558a9eaa5af // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.0
	github.com/prometheus/common v0.9.1
	github.com/securego/gosec v0.0.0-20200302134848-c998389da2ac // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spaceapegames/go-wavefront v1.6.2
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/tommy-muehle/go-mnd v1.3.0 // indirect
	github.com/valyala/fasttemplate v1.0.1
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5 // indirect
	github.com/zorkian/go-datadog-api v2.27.0+incompatible
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/tools v0.0.0-20200306191617-51e69f71924f // indirect
	gonum.org/v1/gonum v0.7.0 // indirect
	gopkg.in/ini.v1 v1.54.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71 // indirect
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/apiserver v0.17.3
	k8s.io/cli-runtime v0.17.3
	k8s.io/client-go v1.5.1
	k8s.io/code-generator v0.17.3
	k8s.io/component-base v0.17.3
	k8s.io/gengo v0.0.0-20200205140755-e0e292d8aa12 // indirect
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe
	k8s.io/kubectl v0.16.4
	k8s.io/kubernetes v1.17.3
	k8s.io/utils v0.0.0-20200229041039-0a110f9eb7ab
	mvdan.cc/unparam v0.0.0-20191111180625-960b1ec0f2c2 // indirect
	sigs.k8s.io/controller-tools v0.2.5 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
	sourcegraph.com/sqs/pbtypes v1.0.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3-beta.0
	k8s.io/apiserver => k8s.io/apiserver v0.17.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.0
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.3
	k8s.io/code-generator => k8s.io/code-generator v0.17.4-beta.0
	k8s.io/component-base => k8s.io/component-base v0.17.3
	k8s.io/cri-api => k8s.io/cri-api v0.17.4-beta.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.3
	k8s.io/kubectl => k8s.io/kubectl v0.17.3
	k8s.io/kubelet => k8s.io/kubelet v0.17.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.3
	k8s.io/metrics => k8s.io/metrics v0.17.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.3
)

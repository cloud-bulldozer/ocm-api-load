module github.com/cloud-bulldozer/ocm-api-load

go 1.15

require (
	github.com/Rican7/retry v0.1.0
	github.com/c2h5oh/datasize v0.0.0-20200825124411-48ed595a09d2 // indirect
	github.com/influxdata/tdigest v0.0.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	github.com/openshift-online/ocm-sdk-go v0.1.165
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/viper v1.6.1
	github.com/tsenart/go-tsz v0.0.0-20180814235614-0bd30b3df1c3 // indirect
	github.com/tsenart/vegeta v11.4.0+incompatible // indirect
	github.com/tsenart/vegeta/v12 v12.8.4
	github.com/zgalor/weberr v0.6.0
	golang.org/x/net v0.0.0-20210421230115-4e50805a0758 // indirect
	k8s.io/apimachinery v0.19.0
)

replace (
	bitbucket.org/ww/goautoneg => github.com/munnerz/goautoneg v0.0.0-20120707110453-a547fc61f48d
	github.com/appscode/jsonpatch => github.com/gomodules/jsonpatch v0.0.0-20190108182946-7c0e3b262f30
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.1.0 // Pin non-versioned import to v22.1.0
	github.com/golang/glog => github.com/kubermatic/glog-logrus v0.0.0-20180829085450-3fa5b9870d1d
	//Temporary fix for openAPI issue in kubernetes version 0.19.0 used by hive 0.12
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	github.com/metal3-io/baremetal-operator => github.com/openshift/baremetal-operator v0.0.0-20200715132148-0f91f62a41fe
	github.com/metal3-io/cluster-api-provider-baremetal => github.com/openshift/cluster-api-provider-baremetal v0.0.0-20190821174549-a2a477909c1d // Pin OpenShift fork
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/terraform-providers/terraform-provider-aws => github.com/openshift/terraform-provider-aws v1.60.1-0.20200526184553-1a716dcc0fa8 // Pin to openshift fork with tag v2.60.0-openshift-1
	github.com/terraform-providers/terraform-provider-azurerm => github.com/openshift/terraform-provider-azurerm v1.41.1-openshift-3 // Pin to openshift fork with IPv6 fixes
	github.com/terraform-providers/terraform-provider-ignition/v2 => github.com/community-terraform-providers/terraform-provider-ignition/v2 v2.1.0
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
	k8s.io/klog => github.com/tzvatot/klog-logrus v1.0.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
	sigs.k8s.io/cluster-api-provider-aws => github.com/openshift/cluster-api-provider-aws v0.2.1-0.20200506073438-9d49428ff837 // Pin OpenShift fork
	sigs.k8s.io/cluster-api-provider-azure => github.com/openshift/cluster-api-provider-azure v0.1.0-alpha.3.0.20200120114645-8a9592f1f87b // Pin OpenShift fork
	sigs.k8s.io/cluster-api-provider-openstack => github.com/openshift/cluster-api-provider-openstack v0.0.0-20200526112135-319a35b2e38e // Pin OpenShift fork
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.4.0
	vbom.ml/util => github.com/fvbommel/util v0.0.0-20180919145318-efcd4e0f9787 // https://github.com/fvbommel/util/issues/6
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm

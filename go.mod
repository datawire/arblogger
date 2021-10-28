module github.com/datawire/arblogger

go 1.15

require (
	github.com/datawire/dlib v1.2.4
	github.com/envoyproxy/go-control-plane v0.9.10-0.20210907150352-cf90f659a021
	github.com/envoyproxy/protoc-gen-validate v0.3.0-java.0.20200609174644-bd816e4522c1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/pkg/errors v0.9.1
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	google.golang.org/grpc v1.41.0
)

// // We need inherit these from github.com/datawire/ambassador.git's go.mod
// replace (
// 	k8s.io/api v0.0.0 => k8s.io/api v0.20.2
// 	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.20.2
// 	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.20.2
// 	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.20.2
// 	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.20.2
// 	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.20.2
// 	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.20.2
// 	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.20.2
// 	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.20.2
// 	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.20.2
// 	k8s.io/component-helpers v0.0.0 => k8s.io/component-helpers v0.20.2
// 	k8s.io/controller-manager v0.0.0 => k8s.io/controller-manager v0.20.2
// 	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.20.2
// 	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.20.2
// 	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.20.2
// 	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.20.2
// 	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.20.2
// 	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.20.2
// 	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.20.2
// 	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.20.2
// 	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.20.2
// 	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.20.2
// 	k8s.io/mount-utils v0.0.0 => k8s.io/mount-utils v0.20.2
// 	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.20.2
// )

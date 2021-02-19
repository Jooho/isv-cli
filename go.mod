module github.com/jooho/isv-cli

go 1.15

require (
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/docker/docker v20.10.2+incompatible // indirect
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	github.com/openshift/library-go v0.0.0-20210106214821-c4d0b9c8d55f
	github.com/openshift/oc v4.2.0-alpha.0+incompatible
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.15.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/apiserver v0.20.2 // indirect
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/component-base v0.20.2
	k8s.io/klog/v2 v2.4.0
	k8s.io/kubectl v0.20.2
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
)

// Fix OCP 4.7
replace (
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190116145627-8841ed7cc6ed
	github.com/openshift/oc v4.2.0-alpha.0+incompatible => github.com/openshift/oc v0.0.0-alpha.0.0.20210119130517-6f8f260853ad

	k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20210108114224-194a87c5b03a
	k8s.io/cli-runtime => github.com/openshift/kubernetes-cli-runtime v0.0.0-20210108114725-2ff6add1e911
	k8s.io/client-go => github.com/openshift/kubernetes-client-go v0.0.0-20210108114446-0829bdd68114
	k8s.io/kubectl => github.com/openshift/kubernetes-kubectl v0.0.0-20210108115031-c0d78c0aeda3
)

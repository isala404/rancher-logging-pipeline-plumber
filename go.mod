module github.com/mrsupiri/logging-pipeline-plumber

go 1.16

require (
	github.com/banzaicloud/logging-operator/pkg/sdk v0.7.3
	github.com/go-logr/logr v0.4.0
	github.com/gorilla/mux v1.7.3
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/rs/cors v1.8.0
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	sigs.k8s.io/controller-runtime v0.8.3
)

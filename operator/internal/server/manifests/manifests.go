package manifests

import (
	_ "embed"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

//go:embed log-output-deployment.yaml
var logOutputDeployment []byte

//go:embed log-output-service.yaml
var logOutputService []byte

func GetLogOutput() (d v1.Deployment, s corev1.Service, err error) {
	d = v1.Deployment{}
	err = yaml.Unmarshal(logOutputDeployment, &d)
	s = corev1.Service{}
	err = yaml.Unmarshal(logOutputService, &s)
	return
}

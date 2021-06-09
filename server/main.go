package main

import (
	"context"
	"io/ioutil"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"sigs.k8s.io/yaml"
)

func main() {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", "./kubeconfig.yaml")

	// creates the in-cluster config
	//config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	//clientset.
	//

	d := v1.Deployment{}
	yamlFile, err := ioutil.ReadFile("log-output.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &d)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	//_, err = clientset.AppsV1().Deployments("default").Create(context.TODO(), &d, metav1.CreateOptions{})
	err = clientset.AppsV1().Deployments("default").Delete(context.TODO(), "log-output", metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
}

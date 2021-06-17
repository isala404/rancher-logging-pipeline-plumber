package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type Operator struct {
	CoreClient *kubernetes.Clientset
}

func (o *Operator) DeployLogOutput(w http.ResponseWriter, r *http.Request) {
	d := v1.Deployment{}
	yamlFile, err := ioutil.ReadFile("log-output.yaml")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("{'error': '%v'}", err)))
		return
	}
	err = yaml.Unmarshal(yamlFile, &d)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("{'error': '%v'}", err)))
		return
	}

	_, err = o.CoreClient.AppsV1().Deployments("default").Create(context.TODO(), &d, metav1.CreateOptions{})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("{'error': %v}", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte("{'data': 'output pod deployed'}"))
}
func (o *Operator) DestroyLogOutput(w http.ResponseWriter, r *http.Request) {
	d := v1.Deployment{}
	yamlFile, err := ioutil.ReadFile("log-output.yaml")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("{'error': '%v'}", err)))
		return
	}
	err = yaml.Unmarshal(yamlFile, &d)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("{'error': '%v'}", err)))
		return
	}

	err = o.CoreClient.AppsV1().Deployments("default").Delete(context.TODO(), "log-output", metav1.DeleteOptions{})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("{'error': %v}", err)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte("{'data': 'output pod destroyed'}"))
}

//go:embed build
var content embed.FS

func clientHandler() http.Handler {
	fsys := fs.FS(content)
	contentStatic, _ := fs.Sub(fsys, "build")
	return http.FileServer(http.FS(contentStatic))
}

func main() {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", "./kubeconfig.yaml")

	// creates the in-cluster config
	// config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	o := Operator{
		CoreClient: clientset,
	}

	r := mux.NewRouter()

	// webapp :=

	// Routes consist of a path and a handler function.
	r.HandleFunc("/log-output/deploy", o.DeployLogOutput).Methods("POST")
	r.HandleFunc("/log-output/destroy", o.DestroyLogOutput).Methods("POST")
	r.PathPrefix("/").Handler(clientHandler())

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", r))

}

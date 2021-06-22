package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebServer struct {
	kubeClient client.Client
	port       string
}

// NewWebServer returns an HTTP server that handles webhooks
func NewWebServer(port string, kubeClient client.Client) *WebServer {

	//// use the current context in kubeconfig
	//config, err := clientcmd.BuildConfigFromFlags("", "kubeconfig.yaml")
	//
	//// creates the in-cluster config
	//// config, err := rest.InClusterConfig()
	//if err != nil {
	//	return nil, err
	//}
	//
	//// create the clientset
	//clientset, err := kubernetes.NewForConfig(config)
	//
	//if err != nil {
	//	return &WebServer{}, err
	//}

	return &WebServer{
		kubeClient: kubeClient,
		port:       port,
	}
}

//go:embed build
var content embed.FS

func clientHandler() http.Handler {
	fsys := fs.FS(content)
	contentStatic, _ := fs.Sub(fsys, "build")
	return http.FileServer(http.FS(contentStatic))
}

func (ws *WebServer) ListenAndServe(stopCh <-chan struct{}) {
	r := mux.NewRouter()

	r.HandleFunc("/log-output/deploy", ws.DeployLogOutput).Methods("POST")
	r.HandleFunc("/log-output/destroy", ws.DestroyLogOutput).Methods("POST")
	r.PathPrefix("/").Handler(clientHandler())

	go func() {
		if err := http.ListenAndServe(ws.port, r); err != http.ErrServerClosed {
			log.Printf("Receiver server crashed: %s", err)
			os.Exit(1)
		}
	}()

	// wait for SIGTERM or SIGINT
	<-stopCh
}

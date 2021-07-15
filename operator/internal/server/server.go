package server

import (
	"embed"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"io/fs"
	"log"
	"net/http"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebServer struct {
	kubeClient client.Client
	port       string
	logger     logr.Logger
}

// NewWebServer returns an HTTP server that handles webhooks
func NewWebServer(port string, kubeClient client.Client) *WebServer {
	logger := ctrl.Log.WithName("web-server")
	return &WebServer{
		kubeClient: kubeClient,
		port:       port,
		logger:     logger,
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
	r.PathPrefix("/k8s").HandlerFunc(ws.ProxyToKubeAPI)
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

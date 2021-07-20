package server

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"io/fs"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebServer struct {
	kubeClient client.Client
	port       string
	logger     logr.Logger
	kubeProxy  KubeProxy
}
type KubeProxy struct {
	endpoint       string
	token          string
	proxyTransport *http.Transport
}

const (
	tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

// NewWebServer returns an HTTP server that handles webhooks
func NewWebServer(port string, kubeClient client.Client) *WebServer {
	logger := ctrl.Log.WithName("web-server")

	// Setup proxy data
	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		logger.V(3).Info("failed to read the token from kubernetes secrets")
	}
	CAData, err := ioutil.ReadFile(rootCAFile)
	if err != nil {
		logger.V(3).Info("failed to read the CA Data from kubernetes secrets")
	}

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(CAData); !ok {
		logger.V(3).Info("failed to append k8s custom CA, using system certs only")
	}

	if err != nil {
		logger.V(3).Info("it seems like operator is not running in side cluster, kube-api proxy will not operate as intended")
	}

	return &WebServer{
		kubeClient: kubeClient,
		port:       port,
		logger:     logger,
		kubeProxy: KubeProxy{
			endpoint: "https://" + net.JoinHostPort(os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")),
			token:    string(token),
			proxyTransport: &http.Transport{TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			}},
		},
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

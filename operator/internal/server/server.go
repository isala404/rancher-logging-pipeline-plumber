package server

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
)

type WebServer struct {
	port      string
	logger    logr.Logger
	kubeProxy KubeProxy
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
func NewWebServer(port string) *WebServer {
	logger := ctrl.Log.WithName("web-server")

	// Setup proxy data
	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		logger.V(-2).Info("failed to read the token from kubernetes secrets")
	}
	CAData, err := ioutil.ReadFile(rootCAFile)
	if err != nil {
		logger.V(-2).Info("failed to read the CA Data from kubernetes secrets")
	}

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(CAData); !ok {
		logger.V(-2).Info("failed to append k8s custom CA, using system certs only")
	}

	if err != nil {
		logger.V(-1).Info("it seems like operator is not running in side cluster, kube-api proxy will not operate as intended")
	}

	return &WebServer{
		port:   port,
		logger: logger,
		kubeProxy: KubeProxy{
			endpoint: "https://" + net.JoinHostPort(os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")),
			token:    string(token),
			proxyTransport: &http.Transport{TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			}},
		},
	}
}

func (ws *WebServer) ListenAndServe(stopCh <-chan struct{}) {
	r := mux.NewRouter()
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

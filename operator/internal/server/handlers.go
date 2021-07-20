package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mrsupiri/rancher-logging-explorer/internal/server/manifests"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func (ws *WebServer) WriteResponse(w http.ResponseWriter, httpRes HTTPResponse) {
	res, jsErr := json.Marshal(httpRes)
	if jsErr != nil {
		ws.logger.Error(jsErr, "failed jsonify HTTPResponse")
	}
	_, err := w.Write(res)
	if err != nil {
		ws.logger.Error(err, "failed to write to ResponseWriter")
	}
}

func (ws *WebServer) ProxyToKubeAPI(res http.ResponseWriter, req *http.Request) {
	targetUrl, _ := url.Parse(ws.kubeProxy.endpoint)

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.Transport = ws.kubeProxy.proxyTransport
	req.URL.Path = "/" + strings.Join(strings.Split(req.URL.Path, "/")[2:], "/")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ws.kubeProxy.token))

	ws.logger.V(1).Info(fmt.Sprintf("proxying %s to %s", req.URL, targetUrl))
	proxy.ServeHTTP(res, req)
}

func (ws *WebServer) DeployLogOutput(w http.ResponseWriter, _ *http.Request) {
	var err error
	ws.logger.V(1).Info("Deploying the log-output")
	w.Header().Set("Content-Type", "application/json")

	d, s, err := manifests.GetLogOutput()
	if err != nil {
		ws.logger.Error(err, "failed fetch log-output manifests")
		w.WriteHeader(500)
		ws.WriteResponse(w, HTTPResponse{Error: &HTTPError{Error: err.Error(), Message: "failed fetch log-output manifests"}})
		return
	}

	err = ws.kubeClient.Create(context.TODO(), &d)
	if err != nil {
		ws.logger.Error(err, "failed to create log-output deployment")
		w.WriteHeader(500)
		ws.WriteResponse(w, HTTPResponse{Error: &HTTPError{Error: err.Error(), Message: "failed to create log-output deployment"}})
		return
	}

	err = ws.kubeClient.Create(context.TODO(), &s)
	if err != nil {
		ws.logger.Error(err, "failed to create log-output service")
		w.WriteHeader(500)
		ws.WriteResponse(w, HTTPResponse{Error: &HTTPError{Error: err.Error(), Message: "failed to create log-output service"}})
		return
	}

	w.WriteHeader(200)
	ws.WriteResponse(w, HTTPResponse{Data: &HTTPData{Message: "log-output deployed successfully"}})
}
func (ws *WebServer) DestroyLogOutput(w http.ResponseWriter, r *http.Request) {
	var err error
	ws.logger.V(1).Info("Destroying the log-output")
	w.Header().Set("Content-Type", "application/json")

	d, s, err := manifests.GetLogOutput()
	if err != nil {
		ws.logger.Error(err, "failed fetch log-output manifests")
		w.WriteHeader(500)
		w.WriteHeader(500)
		ws.WriteResponse(w, HTTPResponse{Error: &HTTPError{Error: err.Error(), Message: "failed fetch log-output manifests"}})
		return
	}

	err = ws.kubeClient.Delete(context.TODO(), &d)
	if err != nil {
		ws.logger.Error(err, "failed to destroy log-output deployment")
		w.WriteHeader(500)
		ws.WriteResponse(w, HTTPResponse{Error: &HTTPError{Error: err.Error(), Message: "failed to destroy log-output service"}})
		return
	}

	err = ws.kubeClient.Delete(context.TODO(), &s)
	if err != nil {
		ws.logger.Error(err, "failed to destroy log-output service")
		w.WriteHeader(500)
		ws.WriteResponse(w, HTTPResponse{Error: &HTTPError{Error: err.Error(), Message: "failed to destroy log-output service"}})
		return
	}

	w.WriteHeader(200)
	ws.WriteResponse(w, HTTPResponse{Data: &HTTPData{Message: "log-output destroyed successfully"}})
}

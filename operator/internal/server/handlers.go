package server

import (
	"context"
	"encoding/json"
	"github.com/mrsupiri/rancher-logging-explorer/internal/server/manifests"
	"net/http"
)

func (ws *WebServer) DeployLogOutput(w http.ResponseWriter, _ *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")

	d, s, err := manifests.GetLogOutput()
	if err != nil {
		w.WriteHeader(500)
		res, _ := json.Marshal(
			HTTPResponse{
				Error: &HTTPError{Error: err.Error(), Message: "Failed fetch log-output manifests"},
			})
		_, _ = w.Write(res)
		return
	}

	err = ws.kubeClient.Create(context.TODO(), &d)
	if err != nil {
		w.WriteHeader(500)
		res, _ := json.Marshal(
			HTTPResponse{
				Error: &HTTPError{Error: err.Error(), Message: "Failed to create log-output deployment"},
			})
		_, _ = w.Write(res)
		return
	}

	err = ws.kubeClient.Create(context.TODO(), &s)
	if err != nil {
		w.WriteHeader(500)
		res, _ := json.Marshal(
			HTTPResponse{
				Error: &HTTPError{Error: err.Error(), Message: "Failed to create log-output service"},
			})
		_, _ = w.Write(res)
		return
	}

	w.WriteHeader(200)
	res, _ := json.Marshal(
		HTTPResponse{
			Data: &HTTPData{Message: "log-output deployed successfully"},
		})
	_, _ = w.Write(res)
}
func (ws *WebServer) DestroyLogOutput(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")

	d, s, err := manifests.GetLogOutput()
	if err != nil {
		w.WriteHeader(500)
		res, _ := json.Marshal(
			HTTPResponse{
				Error: &HTTPError{Error: err.Error(), Message: "Failed fetch log-output manifests"},
			})
		_, _ = w.Write(res)
		return
	}

	err = ws.kubeClient.Delete(context.TODO(), &d)
	if err != nil {
		w.WriteHeader(500)
		res, _ := json.Marshal(
			HTTPResponse{
				Error: &HTTPError{Error: err.Error(), Message: "Failed to destroy log-output deployment"},
			})
		_, _ = w.Write(res)
		return
	}

	err = ws.kubeClient.Delete(context.TODO(), &s)
	if err != nil {
		w.WriteHeader(500)
		res, _ := json.Marshal(
			HTTPResponse{
				Error: &HTTPError{Error: err.Error(), Message: "Failed to destroy log-output service"},
			})
		_, _ = w.Write(res)
		return
	}

	w.WriteHeader(200)
	res, _ := json.Marshal(
		HTTPResponse{
			Data: &HTTPData{Message: "log-output destroyed successfully"},
		})
	_, _ = w.Write(res)
}

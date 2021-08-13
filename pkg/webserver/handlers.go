package webserver

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
)

//go:embed build/*
var content embed.FS

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

func clientHandler() http.Handler {
	handler := fsFunc(func(name string) (fs.File, error) {
		assetPath := path.Join("build", name)

		// If we can't find the asset, return the default index.html
		// content
		f, err := content.Open(assetPath)
		if os.IsNotExist(err) {
			return content.Open("build/index.html")
		}

		// Otherwise assume this is a legitimate request routed
		// correctly
		return f, err
	})

	return http.FileServer(http.FS(handler))
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

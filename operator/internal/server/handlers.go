package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

//go:embed build
var content embed.FS

func clientHandler() http.Handler {
	fsys := fs.FS(content)
	contentStatic, _ := fs.Sub(fsys, "build")
	return http.FileServer(http.FS(contentStatic))
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

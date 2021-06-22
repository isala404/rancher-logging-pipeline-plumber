package server

import (
	"net/http"
)

func (ws *WebServer) DeployLogOutput(w http.ResponseWriter, r *http.Request) {
	//d := v1.Deployment{}
	//yamlFile, err := ioutil.ReadFile("operator/internal/manifests/log-output.yaml")
	//if err != nil {
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(500)
	//	w.Write([]byte(fmt.Sprintf("{'error': '%v'}", err)))
	//	return
	//}
	//err = yaml.Unmarshal(yamlFile, &d)
	//if err != nil {
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(500)
	//	w.Write([]byte(fmt.Sprintf("{'error': '%v'}", err)))
	//	return
	//}
	//
	//_, err = ws.CoreClient.AppsV1().Deployments("default").Create(context.TODO(), &d, metav1.CreateOptions{})
	//if err != nil {
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(500)
	//	w.Write([]byte(fmt.Sprintf("{'error': %v}", err)))
	//	return
	//}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte("{'data': 'output pod deployed'}"))
}
func (ws *WebServer) DestroyLogOutput(w http.ResponseWriter, r *http.Request) {
	//d := v1.Deployment{}
	//yamlFile, err := ioutil.ReadFile("log-output.yaml")
	//if err != nil {
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(500)
	//	w.Write([]byte(fmt.Sprintf("{'error': '%v'}", err)))
	//	return
	//}
	//err = yaml.Unmarshal(yamlFile, &d)
	//if err != nil {
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(500)
	//	w.Write([]byte(fmt.Sprintf("{'error': '%v'}", err)))
	//	return
	//}
	//
	//err = ws.CoreClient.AppsV1().Deployments("default").Delete(context.TODO(), "log-output", metav1.DeleteOptions{})
	//if err != nil {
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(500)
	//	w.Write([]byte(fmt.Sprintf("{'error': %v}", err)))
	//	return
	//}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte("{'data': 'output pod destroyed'}"))
}

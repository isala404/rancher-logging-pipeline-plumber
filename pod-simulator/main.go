package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func echo(w http.ResponseWriter, req *http.Request) {
	var logs []interface{}
	err := json.NewDecoder(req.Body).Decode(&logs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, log := range logs{
		jsonString, err := json.Marshal(log)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println(string(jsonString))
	}
}

func main() {
	var webAddr string
	flag.StringVar(&webAddr, "web-addr", ":9090", "The address http server binds to.")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", echo).Methods("POST")

	err := http.ListenAndServe(webAddr, r)
	if err != nil {
		panic(err)
	}
}

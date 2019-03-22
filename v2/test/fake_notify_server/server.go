package main

import (
	"encoding/json"
	"fmt"
	"github.com/kabukky/httpscerts"
	"log"
	"net/http"
	"strings"
)

type patchData struct {
	Metadata map[string]map[string]string `json:metadata`
	Spec     map[string][]string          `json:spec`
}

func record(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.RequestURI, "/")
	ns := paths[4]
	svc := paths[6]
	method := r.Method

	decoder := json.NewDecoder(r.Body)
	var pd patchData
	err := decoder.Decode(&pd)
	if err != nil {
		log.Output(2, fmt.Sprintf("Received request %v, but faild to decode do patchData struct, since: %v", r.Body, err))
		return
	}

	log.Output(2, fmt.Sprintf("%s %s.%s with anntations[tun_mac:%s, node_ip:%s, cgw_enabled:%s], externalIPs:%v\n", method, ns, svc, pd.Metadata["annotations"]["tun_mac"], pd.Metadata["annotations"]["node_ip"], pd.Metadata["annotations"]["cgw_enabled"], pd.Spec["externalIPs"]))
}

/* From: https://www.kaihag.com/https-and-go/ */
func main() {
	// Check if the cert files are available.
	err := httpscerts.Check("cert.pem", "key.pem")
	// If they are not available, generate new ones.
	if err != nil {
		err = httpscerts.Generate("cert.pem", "key.pem", "0.0.0.0:443")
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}
	http.HandleFunc("/api/v1/namespaces/", record)
	err1 := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil)
	log.Fatal(err1)
}

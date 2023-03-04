package main

import (
	"crypto/tls"
	"gitlab.myshuju.top/heshiying/admission/muta"
	"gitlab.myshuju.top/heshiying/admission/valid"
	"net/http"
)

func main() {

	mux := http.ServeMux{}
	x509KeyPair, err := tls.LoadX509KeyPair("/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key")
	if err != nil {
		return
	}
	mux.Handle("/mutate", muta.NewMutatingAdmissionWebhook())
	mux.Handle("/validate", valid.NewValidateAdmissionWebhook())
	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{x509KeyPair},
	}
	server := http.Server{
		Addr:      ":443",
		Handler:   &mux,
		TLSConfig: &tlsConfig,
	}
	server.ListenAndServe()
}

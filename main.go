package main

import (
	"crypto/tls"
	"gitlab.myshuju.top/heshiying/admission/admissionhooktool"
	"gitlab.myshuju.top/heshiying/admission/muta"
	"gitlab.myshuju.top/heshiying/admission/valid"
	"net/http"
)

func main() {
	admissionhooktool.Log.Info("exec main")
	mux := http.ServeMux{}
	x509KeyPair, err := tls.LoadX509KeyPair("/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key")
	if err != nil {
		admissionhooktool.Log.Error(err, "LoadX509KeyPair err")
		return
	}
	mux.Handle("/mutate", muta.NewMutatingAdmissionWebhook())
	mux.Handle("/validate", valid.NewValidateAdmissionWebhook())
	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{x509KeyPair},
	}
	server := http.Server{
		Addr:      ":1443",
		Handler:   &mux,
		TLSConfig: &tlsConfig,
	}
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		admissionhooktool.Log.Error(err, "ListenAndServeTLS err")
		return
	}
	admissionhooktool.Log.Info("exec end")
}

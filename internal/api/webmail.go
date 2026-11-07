package api

import (
	_ "embed"
	"net/http"
)

//go:embed webmail.html
var webmailHTML []byte

func registerWebmail(mux *http.ServeMux) {
	mux.HandleFunc("GET /mail", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(webmailHTML)
	})
	mux.HandleFunc("GET /mail/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(webmailHTML)
	})
}

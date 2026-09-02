package main

import (
	"net/http"
)

func (webapp *WebApp) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./public/"))
	mux.Handle("GET /public/", fileServer)

	mux.HandleFunc("GET /", webapp.showHomepage)
	mux.HandleFunc("GET /login", webapp.startOIDCFlow)
	mux.HandleFunc("GET /logout", webapp.logout)
	mux.HandleFunc("GET /callback", webapp.handleOIDCCallback)

	return mux
}

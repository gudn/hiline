package main

import (
	_ "embed"
	"net/http"
)

type Pages struct{}

//go:embed pages/timeline.html
var timelineHtml []byte

//go:embed pages/document.html
var documentHtml []byte

func (p Pages) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fallthrough
	case "/timeline":
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(timelineHtml)
	case "/document":
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(documentHtml)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

package main

import (
	_ "embed"
	"net/http"
)

type Pages struct{}

//go:embed pages/timeline.html
var timelineHtml []byte

func (p Pages) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fallthrough
	case "/timeline":
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(timelineHtml)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

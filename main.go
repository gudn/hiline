package main

import (
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
)

var (
	debug = flag.Bool("debug", false, "enable debug logging")
	addr  = flag.String("addr", ":1984", "http address to listen")
	root  = flag.String("root", ".", "path to documents")
)

type loggerResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggerResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func loggerMiddleware(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggerResponseWriter{w, 0}
		inner.ServeHTTP(lrw, r)
		slog.InfoContext(
			r.Context(), "request processed",
			"url", r.URL,
			"method", r.Method,
			"code", lrw.statusCode,
		)
	})
}

func main() {
	flag.Parse()

	if *debug {
		debugLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		}))
		slog.SetDefault(debugLogger)
	}

	if stat, err := os.Stat(*root); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(*root, 0o777)
		if err != nil {
			slog.Error("failed create root", "root", *root, "error", err)
			os.Exit(1)
		}
		slog.Info("create root directory", "root", *root)
	} else if err != nil {
		slog.Error("failed get stat of path", "root", *root, "error", err)
		os.Exit(1)
	} else if !stat.IsDir() {
		slog.Error("path is not directory", "root", *root)
		os.Exit(1)
	}

	groups := NewGroups()
	Scan(*root, groups)

	api := NewApi(groups)
	pages := Pages{}

	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	mux.Handle("/", pages)

	slog.Info("starting listening", "addr", *addr, "root", *root)
	err := http.ListenAndServe(*addr, loggerMiddleware(mux))
	slog.Error("failed serve", "error", err)
}

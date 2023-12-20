package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type Api struct {
	groups *Groups
	mux    *http.ServeMux
}

func NewApi(groups *Groups) Api {
	api := Api{
		groups: groups,
		mux:    http.NewServeMux(),
	}

	api.mux.HandleFunc("/api/group", api.Group)
	api.mux.HandleFunc("/api/group/", api.Group)
	api.mux.HandleFunc("/api/document", api.Document)
	api.mux.HandleFunc("/api/document/", api.Document)

	return api
}

func (a Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a Api) Group(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if cutted, found := strings.CutPrefix(r.URL.Path, "/api/group/"); found && cutted != "" {
		groups := a.groups.Get(cutted)
		if len(groups) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		docs := make([]Document, 0)
		for i := 0; i < len(groups); i++ {
			for dp := range groups[i].Documents {
				doc, err := dp.Read()
				if errors.Is(err, os.ErrNotExist) {
					slog.WarnContext(r.Context(), "document is not exists", "group", groups[i].Id, "docPath", dp)
				} else if err != nil {
					slog.ErrorContext(r.Context(), "failed read document", "group", groups[i].Id, "docPath", dp, "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				docs = append(docs, doc)
			}
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(docs)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed encode documents", "group", cutted, "error", err)
		}
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(a.groups.AllGroups())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed encode all groups", "error", err)
		}
	}
}

func (a Api) Document(w http.ResponseWriter, r *http.Request) {
	if cutted, found := strings.CutPrefix(r.URL.Path, "/api/document/"); found && cutted != "" {
		if !strings.HasSuffix(cutted, ".json") {
			slog.WarnContext(r.Context(), "document without json extension", "path", cutted)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		dp := DocumentPath(cutted)
		switch r.Method {
		case http.MethodGet:
			f, err := dp.Open()
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
				return
			} else if err != nil {
				slog.ErrorContext(r.Context(), "failed open document", "docPath", dp, "error", err)
				w.WriteHeader(http.StatusInsufficientStorage)
				return
			}
			defer f.Close()

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = io.Copy(w, f)
			if err != nil {
				slog.ErrorContext(r.Context(), "failed send document", "docPath", dp, "error", err)
			}
		case http.MethodPost:
			contents, err := io.ReadAll(r.Body)
			if err != nil {
				slog.ErrorContext(r.Context(), "failed read document", "docPath", dp, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var doc Document
			err = json.Unmarshal(contents, &doc)
			if err != nil {
				slog.WarnContext(r.Context(), "failed decode document", "docPath", dp, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if !doc.IsValid() {
				slog.WarnContext(r.Context(), "document is not valid", "docPath", dp)
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}

			err = dp.Write(contents)
			if err != nil {
				slog.ErrorContext(r.Context(), "failed write document", "docPath", dp, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				a.groups.ReplaceDocument(dp, doc)
				w.WriteHeader(http.StatusOK)
			}
		case http.MethodDelete:
			err := os.Remove(dp.String())
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else if err != nil {
				slog.ErrorContext(r.Context(), "failed delete document", "docPath", dp, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				a.groups.DeleteDocument(dp)
				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

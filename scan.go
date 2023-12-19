package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func Scan(root string, groups *Groups) {
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if strings.HasPrefix(filepath.Base(path), ".") {
			slog.Info("ignoring hidden item", "path", path, "isDirectory", d.IsDir())
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && filepath.Ext(path) == ".json" {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				slog.Error("failed compute relative path", "path", path, "error", err)
				return err
			}

			f, err := os.Open(path)
			if err != nil {
				slog.Error("failed open item", "path", path, "error", err)
				return err
			}
			defer f.Close()

			var doc Document
			if err := json.NewDecoder(f).Decode(&doc); err != nil {
				slog.Error("failed decode document", "path", path, "error", err)
				return err
			}

			if !doc.IsValid() {
				slog.Error("read invalid document", "path", path)
				return errors.New("invalid document")
			}

			groups.AddDocument(DocumentPath(rel), doc)
		} else {
			slog.Info("ignore item", "path", path, "isDirectory", d.IsDir())
		}

		return nil
	})
}

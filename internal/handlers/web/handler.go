package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed index.html assets/*
var content embed.FS

type Handler struct {
	assets http.Handler
}

func NewHandler() *Handler {
	assetsFS, err := fs.Sub(content, "assets")
	if err != nil {
		panic(err)
	}
	return &Handler{assets: http.FileServer(http.FS(assetsFS))}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := content.ReadFile("index.html")
	if err != nil {
		http.Error(w, "interface is unavailable", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'self'")
	_, _ = w.Write(data)
}

func (h *Handler) Assets() http.Handler {
	return h.assets
}

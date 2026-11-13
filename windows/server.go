package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed index.html
var indexHTML []byte

const fonteURL = "https://www.limeira.sp.leg.br/vereadores/legislatura_atual"

type app struct {
	dir     string // pasta de dados (%APPDATA%\ObservatorioLimeira)
	token   string
	mu      sync.Mutex
	openURL func(string)
}

func newToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (a *app) dataFile() string { return filepath.Join(a.dir, "dados.json") }

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// só atende a própria janela do programa
	if !strings.HasPrefix(r.Host, "127.0.0.1:") || r.URL.Query().Get("t") != a.token {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	case "/state":
		a.mu.Lock()
		defer a.mu.Unlock()
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<20))
			if err != nil || !json.Valid(body) {
				http.Error(w, "dados inválidos", http.StatusBadRequest)
				return
			}
			tmp := a.dataFile() + ".tmp"
			if err := os.WriteFile(tmp, body, 0o600); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := os.Stat(a.dataFile()); err == nil {
				_ = os.Remove(a.dataFile() + ".bak")
				_ = os.Rename(a.dataFile(), a.dataFile()+".bak")
			}
			if err := os.Rename(tmp, a.dataFile()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte("ok"))
			return
		}
		b, err := os.ReadFile(a.dataFile())
		if err != nil || !json.Valid(b) {
			// arquivo principal ausente ou corrompido: tenta a cópia anterior
			b, err = os.ReadFile(a.dataFile() + ".bak")
			if err != nil || !json.Valid(b) {
				b = nil
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(b)
	case "/open":
		if r.Method == http.MethodPost && a.openURL != nil {
			a.openURL(fonteURL)
		}
		_, _ = w.Write([]byte("ok"))
	default:
		http.NotFound(w, r)
	}
}

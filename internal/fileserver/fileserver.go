package fileserver

import (
	"errors"
	"github.com/charmbracelet/log"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

type Server struct {
	fs     fs.ReadDirFS
	prefix string
}

func New(fs fs.ReadDirFS, path string) *Server {
	return &Server{
		fs,
		path,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fullPath := filepath.Join(s.prefix, r.URL.Path)
	if r.URL.Path[len(r.URL.Path)-1] == '/' {
		fullPath += "/index.html"
	}
	mimeGuess := mime.TypeByExtension(filepath.Ext(fullPath))

	encHeadr := r.Header.Get("Accept-Encoding")
	if encHeadr == "" {
		s.handleFile(w, fullPath, mimeGuess, "")
		return
	}

	encEntries := strings.Split(encHeadr, ",")
	for i, entry := range encEntries {
		encEntries[i] = strings.TrimSpace(entry)
	}

	if slices.Contains(encEntries, "zstd") && s.fileExists(fullPath+".zst") {
		s.handleFile(w, fullPath+".zst", mimeGuess, "zstd")
		return
	}
	if slices.Contains(encEntries, "br") && s.fileExists(fullPath+".br") {
		s.handleFile(w, fullPath+".br", mimeGuess, "br")
		return
	}
	if slices.Contains(encEntries, "gzip") && s.fileExists(fullPath+".gz") {
		s.handleFile(w, fullPath+".gz", mimeGuess, "gzip")
		return
	}
	s.handleFile(w, fullPath, mimeGuess, "")
	return
}

func (s *Server) handleFile(w http.ResponseWriter, path string, mime string, enc string) {
	file, err := s.fs.Open(path)
	if errors.Is(fs.ErrInvalid, err) {
		http.Error(w, "500 bad request", http.StatusBadRequest)
		return
	}
	if errors.Is(fs.ErrInvalid, err) {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		log.Errorf("failed to read file: %v", err)
		return
	}

	w.Header().Set("Content-Type", mime)
	if enc != "" {
		w.Header().Set("Content-Encoding", enc)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Errorf("failed to write HTTP: %v", err)
		return
	}
}

func (s *Server) fileExists(path string) bool {
	_, err := s.fs.Open(path)
	if err != nil {
		return false
	}
	return true
}

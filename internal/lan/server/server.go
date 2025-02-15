package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/0x0FACED/rapid/configs"
)

type FileInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Uploaded bool   `json:"uploaded"`
}

type LANServer struct {
	httpServer *http.Server
	fileList   map[string]FileInfo
	mu         sync.Mutex

	config configs.LANServerConfig
}

func New(cfg configs.LANServerConfig) *LANServer {
	mux := http.NewServeMux()
	server := &LANServer{
		httpServer: &http.Server{
			Addr:    "0.0.0.0:8070",
			Handler: mux,
		},
		fileList: make(map[string]FileInfo),
		config:   cfg,
	}
	server.RegisterHandlers(mux)
	return server
}

func (s *LANServer) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/share", s.handleShare)
	mux.HandleFunc("/api/files", s.handleFiles)
	mux.HandleFunc("/api/download/", s.handleDownload)
}

func (s *LANServer) Start() error {
	fmt.Println("Starting http server")
	return s.httpServer.ListenAndServe()
}

func (s *LANServer) handleShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	fileStat, err := os.Stat(input.Path)
	if err != nil {
		http.Error(w, "File not found", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	id := input.Path
	s.fileList[id] = FileInfo{
		ID:   id,
		Name: input.Name,
		Path: input.Path,
		Size: fileStat.Size(),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "File registered successfully",
		"id":      id,
	})
}

func (s *LANServer) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	files := make([]FileInfo, 0, len(s.fileList))
	for _, file := range s.fileList {
		files = append(files, file)
	}

	json.NewEncoder(w).Encode(files)
}

func (s *LANServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	s.mu.Lock()
	file, exists := s.fileList[id]
	s.mu.Unlock()

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, file.Path)
}

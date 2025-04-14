package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

// StartHTTP initializes and starts the HTTP server for the web interface
func (s *EmailServer) StartHTTP() {
	// Serve the index.html file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		content, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Could not read index.html", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write(content)
	})

	// API to get all emails
	http.HandleFunc("/api/emails", func(w http.ResponseWriter, r *http.Request) {
		emails := s.GetEmails()

		// Sort by timestamp in descending order (newest first)
		sort.Slice(emails, func(i, j int) bool {
			return emails[i].Timestamp.After(emails[j].Timestamp)
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emails)
	})

	// API to get a specific email by ID
	http.HandleFunc("/api/emails/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/emails/")
		if id == "" {
			http.Error(w, "Email ID not specified", http.StatusBadRequest)
			return
		}

		email, found := s.GetEmailByID(id)
		if !found {
			http.Error(w, "Email not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(email)
	})

	// API to delete all emails
	http.HandleFunc("/api/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.ClearEmails()

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status":"ok"}`)
	})

	// Serve static files
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:] // Remove leading slash
		content, err := staticFiles.ReadFile(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Set the appropriate content type
		if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}

		w.Write(content)
	})

	// Start the HTTP server
	go func() {
		log.Printf("Web interface available at http://localhost:%d\n", s.httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.httpPort), nil); err != nil {
			log.Printf("HTTP server error: %v\n", err)
		}
	}()
}

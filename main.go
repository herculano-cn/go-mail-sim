package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Captured Email Representation
type Email struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        []string  `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	HTML      bool      `json:"html"`
	Timestamp time.Time `json:"timestamp"`
}

// EmailServer is the server that manages emails
type EmailServer struct {
	emails    []Email
	mutex     sync.RWMutex
	smtpPort  int
	httpPort  int
	idCounter int
	listener  net.Listener
}

// NewEmailServer creates a new instance of the server
func NewEmailServer(smtpPort, httpPort int) *EmailServer {
	return &EmailServer{
		emails:    make([]Email, 0),
		smtpPort:  smtpPort,
		httpPort:  httpPort,
		idCounter: 1,
	}
}

// StartSMTP starts the SMTP server
func (s *EmailServer) StartSMTP() {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.smtpPort))
	if err != nil {
		log.Fatalf("Error starting SMTP server: %v", err)
	}

	go func() {
		fmt.Printf("SMTP server started on port %d\n", s.smtpPort)
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go s.handleSMTPConnection(conn)
		}
	}()
}

func (s *EmailServer) handleSMTPConnection(conn net.Conn) {
	defer conn.Close()

	// Send greeting
	conn.Write([]byte("220 localhost SMTP server ready\r\n"))

	// Read commands
	scanner := bufio.NewScanner(conn)
	var from string
	var to []string
	var data strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.ToUpper(line), "MAIL FROM:") {
			from = strings.TrimPrefix(line, "MAIL FROM:")
			from = strings.TrimSpace(from)
			conn.Write([]byte("250 Sender OK\r\n"))
		} else if strings.HasPrefix(strings.ToUpper(line), "RCPT TO:") {
			recipient := strings.TrimPrefix(line, "RCPT TO:")
			recipient = strings.TrimSpace(recipient)
			to = append(to, recipient)
			conn.Write([]byte("250 Recipient OK\r\n"))
		} else if strings.HasPrefix(strings.ToUpper(line), "DATA") {
			conn.Write([]byte("354 Start mail input; end with <CRLF>.<CRLF>\r\n"))
			for scanner.Scan() {
				line := scanner.Text()
				if line == "." {
					break
				}
				data.WriteString(line + "\r\n")
			}
			conn.Write([]byte("250 Mail accepted\r\n"))

			// Create and store email
			email := Email{
				ID:        fmt.Sprintf("%d", s.idCounter),
				From:      from,
				To:        to,
				Subject:   "", // You might want to parse this from headers
				Body:      data.String(),
				HTML:      false,
				Timestamp: time.Now(),
			}

			s.mutex.Lock()
			s.emails = append(s.emails, email)
			s.idCounter++
			s.mutex.Unlock()

			from = ""
			to = nil
			data.Reset()
		} else if strings.HasPrefix(strings.ToUpper(line), "QUIT") {
			conn.Write([]byte("221 Bye\r\n"))
			return
		} else {
			conn.Write([]byte("500 Unknown command\r\n"))
		}
	}
}

// StartHTTP starts the HTTP server for the web interface
func (s *EmailServer) StartHTTP() {
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", fs)

	// API to get all emails
	http.HandleFunc("/api/emails", func(w http.ResponseWriter, r *http.Request) {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		// Creates a sorted copy by timestamp in descending order (most recent first)
		emails := make([]Email, len(s.emails))
		copy(emails, s.emails)
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

		s.mutex.RLock()
		defer s.mutex.RUnlock()

		for _, email := range s.emails {
			if email.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(email)
				return
			}
		}

		http.Error(w, "Email not found", http.StatusNotFound)
	})

	// API to delete all emails
	http.HandleFunc("/api/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.mutex.Lock()
		s.emails = make([]Email, 0)
		s.mutex.Unlock()

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "{\"status\":\"ok\"}")
	})

	go func() {
		fmt.Printf("Web interface available at http://localhost:%d\n", s.httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", s.httpPort), nil); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()
}

// Shutdown for the server
func (s *EmailServer) Shutdown() {
	if s.listener != nil {
		s.listener.Close()
	}
}

func main() {
	smtpPort := 1025
	httpPort := 8025

	fmt.Println("Starting GoMailSim - Email Simulator for Development")
	fmt.Println("=====================================================")

	server := NewEmailServer(smtpPort, httpPort)
	server.StartSMTP()
	server.StartHTTP()

	// Keeps the program running
	select {}
}

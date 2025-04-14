package server

import (
	"fmt"
	"sync"
	"time"
)

// Email represents a captured email
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
	emails     []Email
	mutex      sync.RWMutex
	smtpPort   int
	httpPort   int
	idCounter  int
	smtpServer SMTPServer
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

// AddEmail adds a new email to the server
func (s *EmailServer) AddEmail(email Email) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Add ID and timestamp if not present
	if email.ID == "" {
		email.ID = formatID(s.idCounter)
		s.idCounter++
	}

	if email.Timestamp.IsZero() {
		email.Timestamp = time.Now()
	}

	s.emails = append(s.emails, email)
}

// GetEmails returns all emails, sorted by timestamp (newest first)
func (s *EmailServer) GetEmails() []Email {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy of the emails slice
	emails := make([]Email, len(s.emails))
	copy(emails, s.emails)

	return emails
}

// GetEmailByID returns a single email by ID
func (s *EmailServer) GetEmailByID(id string) (Email, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, email := range s.emails {
		if email.ID == id {
			return email, true
		}
	}

	return Email{}, false
}

// ClearEmails removes all emails
func (s *EmailServer) ClearEmails() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.emails = make([]Email, 0)
}

// Shutdown stops all services
func (s *EmailServer) Shutdown() {
	if s.smtpServer != nil {
		s.smtpServer.Close()
	}
}

// Helper function to format ID as a string
func formatID(id int) string {
	return fmt.Sprintf("%d", id)
}

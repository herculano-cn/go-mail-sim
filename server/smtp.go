package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// SMTPServer interface for mocking in tests
type SMTPServer interface {
	ListenAndServe() error
	Close() error
}

// smtpServerImpl is the actual SMTP server implementation
type smtpServerImpl struct {
	listener net.Listener
	port     int
	server   *EmailServer
}

// StartSMTP initializes and starts the SMTP server
func (s *EmailServer) StartSMTP() {
	var err error
	impl := &smtpServerImpl{
		port:   s.smtpPort,
		server: s,
	}

	impl.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.smtpPort))
	if err != nil {
		log.Fatalf("Error starting SMTP server: %v", err)
	}

	s.smtpServer = impl

	go func() {
		log.Printf("SMTP server listening on port %d\n", s.smtpPort)
		for {
			conn, err := impl.listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go impl.handleConnection(conn)
		}
	}()
}

func (s *smtpServerImpl) handleConnection(conn net.Conn) {
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
		cmd := strings.ToUpper(line)

		if strings.HasPrefix(cmd, "MAIL FROM:") {
			from = strings.TrimPrefix(line, "MAIL FROM:")
			from = strings.TrimSpace(from)
			conn.Write([]byte("250 Sender OK\r\n"))
		} else if strings.HasPrefix(cmd, "RCPT TO:") {
			recipient := strings.TrimPrefix(line, "RCPT TO:")
			recipient = strings.TrimSpace(recipient)
			to = append(to, recipient)
			conn.Write([]byte("250 Recipient OK\r\n"))
		} else if cmd == "DATA" {
			conn.Write([]byte("354 Start mail input; end with <CRLF>.<CRLF>\r\n"))

			// Read the email data until we see a line with just a period
			for scanner.Scan() {
				line := scanner.Text()
				if line == "." {
					break
				}
				// Remove the leading dot if present (SMTP data escaping)
				if strings.HasPrefix(line, ".") {
					line = line[1:]
				}
				data.WriteString(line + "\r\n")
			}

			conn.Write([]byte("250 Mail accepted\r\n"))

			// Extract subject and body from email data
			content := data.String()
			subject := ""
			body := content
			isHTML := false

			// Process headers and body
			parts := strings.Split(content, "\r\n\r\n")
			if len(parts) >= 2 {
				headers := parts[0]
				body = strings.Join(parts[1:], "\r\n\r\n")

				// Extract subject from headers
				for _, line := range strings.Split(headers, "\r\n") {
					if strings.HasPrefix(strings.ToLower(line), "subject:") {
						subject = strings.TrimSpace(line[8:])
					}
					if strings.Contains(strings.ToLower(line), "content-type: text/html") {
						isHTML = true
					}
				}
			}

			// Create a new email
			email := Email{
				From:      from,
				To:        to,
				Subject:   subject,
				Body:      body,
				HTML:      isHTML,
				Timestamp: time.Now(),
			}

			// Add the email to our server
			s.server.AddEmail(email)

			log.Printf("Email received: %s -> %v\n", from, to)

			from = ""
			to = nil
			data.Reset()
		} else if cmd == "QUIT" {
			conn.Write([]byte("221 Bye\r\n"))
			return
		} else if cmd == "HELO" || cmd == "EHLO" {
			// Handle HELO/EHLO commands
			conn.Write([]byte("250 Hello\r\n"))
		} else if cmd == "RSET" {
			// Reset the current transaction
			from = ""
			to = nil
			data.Reset()
			conn.Write([]byte("250 OK\r\n"))
		} else if cmd == "NOOP" {
			// Do nothing
			conn.Write([]byte("250 OK\r\n"))
		} else {
			// Log unknown commands for debugging
			log.Printf("Unknown SMTP command: %s", line)
			conn.Write([]byte("500 Unknown command\r\n"))
		}
	}
}

// ListenAndServe implements the SMTPServer interface
func (s *smtpServerImpl) ListenAndServe() error {
	// This is a no-op since we already started the server in StartSMTP
	return nil
}

// Close implements the SMTPServer interface
func (s *smtpServerImpl) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

# GoMailSim

GoMailSim is an SMTP server simulator for development, inspired by Mailpit. It allows you to capture emails sent by your application without sending them to real recipients, providing a web interface for viewing the captured emails.

![GoMailSim Screenshot](https://via.placeholder.com/800x450)

## Features

- Lightweight SMTP server on port 1025 (configurable)
- Clean web interface on port 8025 (configurable)
- Captures and displays email metadata (sender, recipient, subject, etc.)
- Support for both HTML and plain text content
- Sorts emails by received date
- Option to clear all emails
- Zero external dependencies (uses only Go standard library)

## Installation

### Via Go Install

```bash
go install github.com/herculano-cn/go-mail-sim/cmd/gomailsim@latest
```

### From Source

```bash
git clone https://github.com/herculano-cn/go-mail-sim.git
cd go-mail-sim
go build -o gomailsim cmd/gomailsim/main.go
```

## Usage

### Running the Server

```bash
gomailsim
```

By default, GoMailSim starts an SMTP server on port 1025 and a web interface on port 8025. You can access the web interface at http://localhost:8025.

### Configuration Options

You can configure the SMTP and HTTP ports using command-line flags:

```bash
gomailsim -smtp=2025 -http=9025
```

### As a Package

You can also use GoMailSim as a package in your Go applications:

```go
package main

import (
    "github.com/herculano-cn/go-mail-sim/server"
    "time"
)

func main() {
    // Create a new email server with SMTP on port 1025 and HTTP on port 8025
    mailServer := server.NewEmailServer(1025, 8025)
    
    // Start the servers
    mailServer.StartSMTP()
    mailServer.StartHTTP()
    
    // Keep your application running
    select {}
    
    // Or stop the server when needed
    // mailServer.Shutdown()
}
```

## Sending Emails to GoMailSim

### Using Go's smtp Package

```go
package main

import (
    "fmt"
    "net/smtp"
)

func main() {
    // Connect to GoMailSim server
    host := "localhost:1025"
    
    // Email details
    from := "sender@example.com"
    to := []string{"recipient@example.com"}
    
    // Email content
    subject := "GoMailSim Test"
    body := "This is a test email sent to GoMailSim!"
    
    // Build the email
    message := []byte("Subject: " + subject + "\r\n" +
                      "From: " + from + "\r\n" +
                      "To: " + to[0] + "\r\n" +
                      "Content-Type: text/plain; charset=UTF-8\r\n" +
                      "\r\n" +
                      body)
    
    // Send the email
    err := smtp.SendMail(host, nil, from, to, message)
    if err != nil {
        fmt.Println("Error sending email:", err)
        return
    }
    
    fmt.Println("Email sent successfully!")
}
```

### Using HTML Content

```go
message := []byte("Subject: HTML Test\r\n" +
                 "From: sender@example.com\r\n" +
                 "To: recipient@example.com\r\n" +
                 "Content-Type: text/html; charset=UTF-8\r\n" +
                 "\r\n" +
                 "<html><body><h1>Hello World!</h1><p>This is an <b>HTML</b> email.</p></body></html>")
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- Inspired by [Mailpit](https://github.com/axllent/mailpit) and similar projects
- Built with Go standard library
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/herculano-cn/go-mail-sim/server"
)

func main() {
	// Parse command line flags
	smtpPort := flag.Int("smtp", 1025, "SMTP server port")
	httpPort := flag.Int("http", 8025, "HTTP server port")
	flag.Parse()

	fmt.Println("Starting GoMailSim - Email Testing Server")
	fmt.Println("=========================================")
	fmt.Printf("SMTP server on port: %d\n", *smtpPort)
	fmt.Printf("Web interface on:    http://localhost:%d\n\n", *httpPort)

	// Create and start the email server
	emailServer := server.NewEmailServer(*smtpPort, *httpPort)
	emailServer.StartSMTP()
	emailServer.StartHTTP()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down GoMailSim...")
	emailServer.Shutdown()
	fmt.Println("Goodbye!")
}

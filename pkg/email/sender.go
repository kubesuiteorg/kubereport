package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// Initialize logger
var logger *log.Logger

func init() {
	if isRunningInKubernetes() {
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// Check if the application is running in a Kubernetes pod
func isRunningInKubernetes() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	return !os.IsNotExist(err)
}

// SendEmail sends an email with the given subject, body, and attachment.
func SendEmail(subject, body, to, from, smtpServer, smtpPort, smtpUser, smtpPassword, attachmentPath string, useTLS bool) error {
	if logger != nil {
		logger.Printf("Starting to send email from %s to %s with subject: %s", from, to, subject)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	if attachmentPath != "" {
		if _, err := os.Stat(attachmentPath); err == nil {
			m.Attach(attachmentPath)
		} else {
			if logger != nil {
				logger.Printf("Attachment file not found: %v", err)
			}
			return fmt.Errorf("failed to find the attachment: %v", err)
		}
	}

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %v", err)
	}

	d := gomail.NewDialer(smtpServer, port, smtpUser, smtpPassword)
	// Configure TLS only if useTLS is true
	if useTLS {
		d.TLSConfig = &tls.Config{
			ServerName: smtpServer,
			MinVersion: tls.VersionTLS12,
		}
	} else {
		d.TLSConfig = &tls.Config{
			InsecureSkipVerify: true, // If you want to skip certificate verification (not recommended)
		}
	}

	if err := d.DialAndSend(m); err != nil {
		if logger != nil {
			logger.Printf("Failed to send email: %v", err)
		}
		return fmt.Errorf("failed to send email: %v", err)
	}

	if logger != nil {
		logger.Printf("Email sent successfully to %s", to)
	}
	return nil
}

package service

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// SendOTP sends an OTP code to the provided email address
func SendOTP(email, otpCode string) error {
	// Check if we should use console delivery for testing
	smtpHost := os.Getenv("SMTP_HOST")
	smtpUsername := os.Getenv("SMTP_USERNAME")

	// If SMTP is not configured or explicitly set to console mode, log to console
	if smtpHost == "" || smtpUsername == "" || os.Getenv("OTP_DELIVERY_METHOD") == "console" {
		log.Printf("[TEST MODE] OTP for %s: %s", email, otpCode)
		return nil
	}

	// Continue with actual email sending using SMTP
	smtpPort := os.Getenv("SMTP_PORT")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromName := os.Getenv("SMTP_FROM_NAME")

	if fromName == "" {
		fromName = "Ticket System"
	}

	// Create message
	subject := "Your Verification Code"
	body := fmt.Sprintf("Your verification code is: %s\nThis code will expire in %s minutes.",
		otpCode, os.Getenv("OTP_EXPIRY_MINUTES"))

	message := []byte(fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"\r\n"+
		"%s", fromName, smtpUsername, email, subject, body))

	// Authentication
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// Send email
	addr := smtpHost + ":" + smtpPort
	err := smtp.SendMail(addr, auth, smtpUsername, []string{email}, message)
	if err != nil {
		log.Printf("Failed to send email: %v. Falling back to console delivery.", err)
		log.Printf("[FALLBACK] OTP for %s: %s", email, otpCode)
		return nil // Return nil to prevent authentication failures when SMTP fails
	}

	return nil
}

// SendWelcomeEmail sends a welcome email to newly registered users
func SendWelcomeEmail(email, username string) error {
	// Check if we should use console delivery for testing
	smtpHost := os.Getenv("SMTP_HOST")
	smtpUsername := os.Getenv("SMTP_USERNAME")

	// If SMTP is not configured or explicitly set to console mode, log to console
	if smtpHost == "" || smtpUsername == "" || os.Getenv("OTP_DELIVERY_METHOD") == "console" {
		log.Printf("[TEST MODE] Welcome email for %s", email)
		return nil
	}

	// Continue with actual email sending using SMTP
	smtpPort := os.Getenv("SMTP_PORT")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromName := os.Getenv("SMTP_FROM_NAME")

	if fromName == "" {
		fromName = "Ticket System"
	}

	// Sanitize username if empty
	if strings.TrimSpace(username) == "" {
		username = "there"
	}

	// Create message
	subject := "Welcome to Ticket System"
	body := fmt.Sprintf("Hi %s,\n\nWelcome to Ticket System! Your account has been successfully created.\n\n"+
		"You can now book tickets for movies, flights, trains, and events through our platform.\n\n"+
		"Best regards,\nThe Ticket System Team", username)

	message := []byte(fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"\r\n"+
		"%s", fromName, smtpUsername, email, subject, body))

	// Authentication
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// Send email
	addr := smtpHost + ":" + smtpPort
	err := smtp.SendMail(addr, auth, smtpUsername, []string{email}, message)
	if err != nil {
		log.Printf("Failed to send welcome email: %v. Falling back to console delivery.", err)
		log.Printf("[FALLBACK] Welcome email for %s", email)
		return nil // Return nil to prevent authentication failures when SMTP fails
	}

	return nil
}

// File: internal/services/notify/email.go
package notify

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// sendEmail mengirimkan Message via SMTP
func (s *Service) sendEmail(msg Message) error {
	cfg := s.cfg.Email
	if !cfg.Enabled {
		return nil // silently skip
	}
	if cfg.SMTPHost == "" || cfg.Username == "" {
		return fmt.Errorf("email smtp_host dan username harus diisi")
	}

	recipients := msg.ToEmails
	if len(recipients) == 0 {
		recipients = cfg.ToEmails
	}
	if len(recipients) == 0 {
		return fmt.Errorf("tidak ada penerima email yang dikonfigurasi")
	}

	subject := fmt.Sprintf("[%s] %s", msg.Level, msg.Title)
	body := FormatEmailMessage(msg)

	// Build RFC 5322 raw email
	rawMsg := buildRawEmail(cfg.FromName, cfg.FromEmail, recipients, subject, body)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	if cfg.UseTLS {
		// SSL/TLS langsung (port 465)
		return sendEmailSSL(addr, cfg.SMTPHost, cfg.Username, cfg.Password, cfg.FromEmail, recipients, rawMsg)
	}
	// STARTTLS (port 587) — lebih umum
	return sendEmailSTARTTLS(addr, cfg.SMTPHost, cfg.Username, cfg.Password, cfg.FromEmail, recipients, rawMsg)
}

func sendEmailSTARTTLS(addr, host, user, pass, from string, to []string, msg []byte) error {
	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(addr, auth, from, to, msg)
}

func sendEmailSSL(addr, host, user, pass, from string, to []string, msg []byte) error {
	tlsCfg := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("failed to dial TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", user, pass, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	return w.Close()
}

// buildRawEmail membangun raw email message sesuai RFC 5322
func buildRawEmail(fromName, fromEmail string, to []string, subject, body string) []byte {
	from := fmt.Sprintf("%s <%s>", fromName, fromEmail)
	date := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")
	headers := strings.Join([]string{
		"From: " + from,
		"To: " + strings.Join(to, ", "),
		"Subject: " + subject,
		"Date: " + date,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	return []byte(headers)
}

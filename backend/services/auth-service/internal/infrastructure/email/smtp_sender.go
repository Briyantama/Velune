package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SMTPOtpSender delivers OTP codes via SMTP.
// It must never log OTP secrets.
type SMTPOtpSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	TLS      bool
	Log      *zap.Logger
}

func NewSMTPOtpSender(host, port, username, password, from string, tls bool, log *zap.Logger) *SMTPOtpSender {
	return &SMTPOtpSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		TLS:      tls,
		Log:      log,
	}
}

func (s *SMTPOtpSender) SendOTP(ctx context.Context, toEmail string, otpCode string, expiresAt time.Time) error {
	if s.Host == "" || s.Port == "" || s.From == "" {
		return fmt.Errorf("smtp otp sender not configured")
	}
	if toEmail == "" {
		return fmt.Errorf("toEmail is required")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	addr := net.JoinHostPort(s.Host, s.Port)

	// RFC 5322-ish minimal message.
	subject := "Velune: your verification code"
	body := fmt.Sprintf(
		"Your one-time verification code is %s.\n\nThis code expires at %s (UTC).\n\nIf you did not request this, you can ignore this email.",
		otpCode,
		expiresAt.Format(time.RFC3339),
	)

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", s.From),
		fmt.Sprintf("To: %s", toEmail),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.Username != "" || s.Password != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}

	// Use Dial + SMTP client so we can control STARTTLS based on configuration.
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.Close()
	}()

	if s.TLS {
		if err := c.StartTLS(tlsConfig()); err != nil {
			return err
		}
	}
	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return err
		}
	}

	if err := c.Mail(s.From); err != nil {
		return err
	}
	if err := c.Rcpt(toEmail); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, _ = w.Write([]byte(msg))
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// tlsConfig returns a minimal TLS config for StartTLS.
func tlsConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
}

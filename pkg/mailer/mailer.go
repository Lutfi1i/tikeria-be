package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"

	"tikeria-be/config"
)

type Mailer interface {
	SendOTPEmail(recipientEmail, recipientName, otpCode string) error
}

type smtpMailer struct {
	cfg *config.AppConfig
}

func NewMailer(cfg *config.AppConfig) Mailer {
	return &smtpMailer{cfg: cfg}
}

func (m *smtpMailer) SendOTPEmail(recipientEmail, recipientName, otpCode string) error {
	if m.cfg == nil || strings.TrimSpace(m.cfg.SMTPHost) == "" {
		log.Printf("[DEV/MOCK EMAIL] OTP for %s (%s): %s (Valid for %d minutes)",
			recipientEmail, recipientName, otpCode, m.cfg.OTPExpiryMinutes)
		return nil
	}

	htmlBody, textBody, err := RenderOTPEmail(OTPTemplateData{
		RecipientName: recipientName,
		OTPCode:       otpCode,
		ExpiryMinutes: m.cfg.OTPExpiryMinutes,
		SupportEmail:  "support@tikeria.id",
	})
	if err != nil {
		return err
	}

	from := m.cfg.SMTPFrom
	if from == "" {
		from = "no-reply@tikeria.id"
	}
	subject := fmt.Sprintf("Kode OTP Tikeria: %s", otpCode)

	boundary := fmt.Sprintf("=_tikeria_%d", time.Now().UnixNano())

	var msg bytesBuffer
	msg.WriteString(fmt.Sprintf("From: Tikeria <%s>\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipientEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// Plain text part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n\r\n")

	// HTML part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n\r\n")

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	addr := fmt.Sprintf("%s:%d", m.cfg.SMTPHost, m.cfg.SMTPPort)

	// Direct SSL/TLS on port 465
	if m.cfg.SMTPPort == 465 {
		tlsConfig := &tls.Config{
			ServerName: m.cfg.SMTPHost,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP SSL (port 465): %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, m.cfg.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		if m.cfg.SMTPUser != "" && m.cfg.SMTPPassword != "" {
			auth := smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPassword, m.cfg.SMTPHost)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP auth error: %w", err)
			}
		}

		if err := client.Mail(from); err != nil {
			return err
		}
		if err := client.Rcpt(recipientEmail); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg.Bytes())
		if err != nil {
			return err
		}
		return w.Close()
	}

	// Standard SMTP with STARTTLS (Port 587 or 25)
	var auth smtp.Auth
	if m.cfg.SMTPUser != "" && m.cfg.SMTPPassword != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPassword, m.cfg.SMTPHost)
	}

	// Try standard SendMail with timeout dialer
	err = sendMailWithStartTLS(addr, auth, from, []string{recipientEmail}, msg.Bytes(), m.cfg.SMTPHost)
	if err != nil {
		log.Printf("[SMTP ERROR] Failed to send email to %s: %v", recipientEmail, err)
		return err
	}

	return nil
}

type bytesBuffer struct {
	data []byte
}

func (b *bytesBuffer) WriteString(s string) {
	b.data = append(b.data, s...)
}

func (b *bytesBuffer) Bytes() []byte {
	return b.data
}

func sendMailWithStartTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: host}
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	return w.Close()
}

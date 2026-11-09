package mailer

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"os"
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
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	return &smtpMailer{cfg: cfg}
}

func (m *smtpMailer) SendOTPEmail(recipientEmail, recipientName, otpCode string) error {
	if m.cfg == nil || strings.TrimSpace(m.cfg.SMTPHost) == "" {
		log.Printf("[DEV/MOCK EMAIL] OTP for %s (%s): %s (Valid for %d minutes)",
			recipientEmail, recipientName, otpCode, m.cfg.OTPExpiryMinutes)
		return nil
	}

	assetBaseURL := strings.TrimSpace(os.Getenv("EMAIL_ASSET_BASE_URL"))
	if assetBaseURL == "" {
		// Default to high-speed public CDN so images load inline without triggering email attachments
		assetBaseURL = "https://cdn.jsdelivr.net/gh/Lutfi1i/tikeria-be@main/pkg/mailer/assets/"
	}

	useCID := assetBaseURL == "cid:"
	assetURL := assetBaseURL
	if !useCID && !strings.HasSuffix(assetURL, "/") {
		assetURL += "/"
	}

	htmlBody, textBody, err := RenderOTPEmail(OTPTemplateData{
		RecipientName: recipientName,
		OTPCode:       otpCode,
		ExpiryMinutes: m.cfg.OTPExpiryMinutes,
		AssetURL:      assetURL,
		InstagramURL:  "https://www.instagram.com/tikeria.id",
		FacebookURL:   "https://www.facebook.com/tikeria.id",
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

	var msg bytesBuffer
	msg.WriteString(fmt.Sprintf("From: Tikeria <%s>\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipientEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	relatedBoundary := fmt.Sprintf("=_tikeria_rel_%d", time.Now().UnixNano())
	altBoundary := fmt.Sprintf("=_tikeria_alt_%d", time.Now().UnixNano())

	encodeQP := func(s string) string {
		var buf bytes.Buffer
		w := quotedprintable.NewWriter(&buf)
		_, _ = w.Write([]byte(s))
		_ = w.Close()
		return buf.String()
	}

	if useCID {
		// When CID mode is requested: use multipart/related with explicit type
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/related; type=\"multipart/alternative\"; boundary=\"%s\"\r\n\r\n", relatedBoundary))

		// Nested multipart/alternative for text and html
		msg.WriteString(fmt.Sprintf("--%s\r\n", relatedBoundary))
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))

		// Plain text part
		msg.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		msg.WriteString(encodeQP(textBody))
		msg.WriteString("\r\n")

		// HTML part
		msg.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		msg.WriteString(encodeQP(htmlBody))
		msg.WriteString("\r\n")

		msg.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))

		// Attach inline images
		imageNames := []string{"mark.png", "stripes.png", "mail.png", "confetti.png", "instagram.png", "facebook.png"}
		for _, name := range imageNames {
			data, err := GetEmbeddedAsset(name)
			if err != nil {
				continue
			}
			b64 := base64.StdEncoding.EncodeToString(data)

			msg.WriteString(fmt.Sprintf("--%s\r\n", relatedBoundary))
			msg.WriteString(fmt.Sprintf("Content-Type: image/png; name=\"%s\"\r\n", name))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n")
			msg.WriteString(fmt.Sprintf("Content-ID: <%s@tikeria.id>\r\n", name))
			msg.WriteString("Content-Disposition: inline\r\n\r\n")

			for i := 0; i < len(b64); i += 76 {
				end := i + 76
				if end > len(b64) {
					end = len(b64)
				}
				msg.WriteString(b64[i:end] + "\r\n")
			}
		}

		msg.WriteString(fmt.Sprintf("--%s--\r\n", relatedBoundary))
	} else {
		// Production standard: multipart/alternative without file attachments
		// Images load directly via CDN inside the HTML body without cluttering email attachments
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))

		msg.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		msg.WriteString(encodeQP(textBody))
		msg.WriteString("\r\n")

		msg.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		msg.WriteString(encodeQP(htmlBody))
		msg.WriteString("\r\n")

		msg.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
	}

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

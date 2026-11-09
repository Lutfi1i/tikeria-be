package mailer

import (
	"bytes"
	"fmt"
	"html/template"
)

type OTPTemplateData struct {
	RecipientName string
	OTPCode       string
	ExpiryMinutes int
	SupportEmail  string
}

const otpHTMLTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Kode Verifikasi Tikeria</title>
</head>
<body style="margin: 0; padding: 0; background-color: #F8FAFC; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #1E293B;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color: #F8FAFC; padding: 32px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width: 520px; background-color: #FFFFFF; border: 1px solid #E2E8F0; border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.05);">
          <!-- Header -->
          <tr>
            <td style="padding: 28px 32px 20px 32px; border-bottom: 1px solid #F1F5F9;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td>
                    <span style="font-size: 20px; font-weight: 800; letter-spacing: -0.5px; color: #0F172A;">TIKERIA</span>
                  </td>
                  <td align="right">
                    <span style="font-size: 12px; font-weight: 600; color: #64748B; text-transform: uppercase; letter-spacing: 0.5px;">Verifikasi Email</span>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding: 32px 32px 24px 32px;">
              <p style="margin: 0 0 16px 0; font-size: 15px; line-height: 24px; color: #334155;">
                Halo{{if .RecipientName}} <strong>{{.RecipientName}}</strong>{{end}},
              </p>
              <p style="margin: 0 0 24px 0; font-size: 15px; line-height: 24px; color: #334155;">
                Berikut adalah kode verifikasi (OTP) untuk menyelesaikan pendaftaran akun Anda di Tikeria:
              </p>
              <!-- OTP Box -->
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin: 0 0 24px 0;">
                <tr>
                  <td align="center" style="background-color: #F1F5F9; border: 1px solid #CBD5E1; border-radius: 8px; padding: 18px 24px;">
                    <span style="font-family: 'SF Mono', Consolas, 'Liberation Mono', Menlo, monospace; font-size: 32px; font-weight: 700; letter-spacing: 8px; color: #0F172A; display: inline-block;">
                      {{.OTPCode}}
                    </span>
                  </td>
                </tr>
              </table>
              <p style="margin: 0 0 16px 0; font-size: 13px; line-height: 20px; color: #64748B;">
                Kode ini berlaku selama <strong>{{.ExpiryMinutes}} menit</strong>. Demi keamanan akun Anda, jangan berikan kode ini kepada siapa pun, termasuk pihak Tikeria.
              </p>
              <p style="margin: 0; font-size: 13px; line-height: 20px; color: #64748B;">
                Jika Anda tidak merasa mendaftar di Tikeria, abaikan email ini.
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding: 20px 32px; background-color: #F8FAFC; border-top: 1px solid #E2E8F0; font-size: 12px; line-height: 18px; color: #94A3B8; text-align: center;">
              Tikeria &bull; Platform Tiket & Event Indonesia<br>
              Butuh bantuan? Hubungi <a href="mailto:{{.SupportEmail}}" style="color: #2563EB; text-decoration: none;">{{.SupportEmail}}</a>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

func RenderOTPEmail(data OTPTemplateData) (htmlBody string, plainTextBody string, err error) {
	if data.ExpiryMinutes <= 0 {
		data.ExpiryMinutes = 10
	}
	if data.SupportEmail == "" {
		data.SupportEmail = "support@tikeria.id"
	}

	tmpl, err := template.New("otp_email").Parse(otpHTMLTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse OTP email template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("failed to render OTP email template: %w", err)
	}

	plainText := fmt.Sprintf(
		"Halo%s,\n\nKode verifikasi Tikeria Anda adalah: %s\n\nKode ini berlaku selama %d menit.\nDemi keamanan akun Anda, jangan berikan kode ini kepada siapa pun.\n\nJika Anda tidak merasa mendaftar di Tikeria, abaikan email ini.\n\n--\nTikeria Support: %s",
		func() string {
			if data.RecipientName != "" {
				return " " + data.RecipientName
			}
			return ""
		}(),
		data.OTPCode,
		data.ExpiryMinutes,
		data.SupportEmail,
	)

	return buf.String(), plainText, nil
}

package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed assets/*
var emailAssets embed.FS

type OTPTemplateData struct {
	RecipientName string
	OTPCode       string
	ExpiryMinutes int
	AssetURL      string
	InstagramURL  string
	FacebookURL   string
	SupportEmail  string
}

const otpHTMLTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="color-scheme" content="light">
  <meta name="supported-color-schemes" content="light">
  <title>Verifikasi Email Tikeria</title>
  <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;600;700;800&display=swap" rel="stylesheet">
</head>
<body style="margin:0;padding:0;background-color:#F2F2F4;">

  <!-- Preheader -->
  <div style="display:none;max-height:0;overflow:hidden;opacity:0;font-size:1px;line-height:1px;color:#F2F2F4;">
    Kode verifikasi Tikeria kamu: {{.OTPCode}}. Berlaku {{.ExpiryMinutes}} menit.
  </div>

  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#F2F2F4">
    <tr>
      <td align="center" style="padding:24px 12px;">

        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="max-width:480px;background-color:#FFFFFF;">

          <!-- Header -->
          <tr>
            <td bgcolor="#FF2056" style="background-color:#FF2056;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td width="1" valign="middle" style="padding:8px 0 8px 20px;">
                    <table role="presentation" cellpadding="0" cellspacing="0" border="0" bgcolor="#FFFFFF" style="background-color:#FFFFFF;">
                      <tr>
                        <td valign="middle" style="padding:3px 0 3px 8px;line-height:0;font-size:0;">
                          <img src="{{.AssetURL}}mark.png" width="42" height="34" alt="" style="display:block;width:42px;height:34px;border:0;">
                        </td>
                        <td valign="middle" style="padding:0 8px 0 4px;font-family:'Poppins','Helvetica Neue',Arial,sans-serif;font-size:29px;line-height:40px;font-weight:800;letter-spacing:-0.5px;color:#0B0B0F;white-space:nowrap;">TIKE</td>
                      </tr>
                    </table>
                  </td>
                  <td width="1" valign="middle" style="padding-left:5px;font-family:'Poppins','Helvetica Neue',Arial,sans-serif;font-size:29px;line-height:40px;font-weight:800;letter-spacing:-0.5px;color:#FFFFFF;white-space:nowrap;">RIA</td>
                  <td align="right" valign="top" style="line-height:0;font-size:0;">
                    <img src="{{.AssetURL}}stripes.png" width="100" height="56" alt="" style="display:block;margin-left:auto;border:0;width:100px;height:56px;">
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Judul -->
          <tr>
            <td style="padding:28px 24px 0 24px;font-family:'Poppins','Helvetica Neue',Arial,sans-serif;">
              <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td style="padding-right:10px;line-height:0;">
                    <img src="{{.AssetURL}}mail.png" width="22" height="16" alt="" style="display:block;border:0;">
                  </td>
                  <td style="font-size:19px;line-height:26px;font-weight:700;color:#0B0B0F;">Verifikasi Email</td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Sapaan -->
          <tr>
            <td style="padding:20px 24px 0 24px;font-family:'Poppins','Helvetica Neue',Arial,sans-serif;font-size:16px;line-height:26px;color:#1A1A1F;">
              Halo <strong style="font-weight:700;color:#0B0B0F;">{{.RecipientName}}</strong>,<br>
              Berikut adalah kode verifikasi (OTP) untuk menyelesaikan pendaftaran akun Anda di Tikeria:
            </td>
          </tr>

          <!-- Kode OTP -->
          <tr>
            <td style="padding:24px 24px 0 24px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td align="center" style="border:1px solid #D8D8DE;border-radius:12px;padding:26px 8px;font-family:'Poppins','Helvetica Neue',Arial,sans-serif;font-size:46px;line-height:52px;font-weight:800;letter-spacing:16px;color:#0B0B0F;">
                    <span style="padding-left:16px;">{{.OTPCode}}</span>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Masa berlaku + keamanan -->
          <tr>
            <td style="padding:16px 24px 0 24px;font-family:'Poppins','Helvetica Neue',Arial,sans-serif;font-size:13px;line-height:21px;color:#6B6B76;">
              Kode berlaku <strong style="color:#0B0B0F;">{{.ExpiryMinutes}} menit</strong>. Jangan bagikan kode ini kepada siapa pun, termasuk pihak yang mengaku dari Tikeria.
            </td>
          </tr>

          <!-- Sosial Media Icons (Instagram & Facebook with clickable links) -->
          <tr>
            <td align="center" style="padding:28px 24px 0 24px;">
              <table role="presentation" cellpadding="0" cellspacing="0" border="0" align="center">
                <tr>
                  <td align="center" style="padding:0 12px;">
                    <a href="{{.InstagramURL}}" target="_blank" rel="noopener noreferrer" style="display:inline-block;text-decoration:none;">
                      <img src="{{.AssetURL}}instagram.png" width="28" height="28" alt="Instagram" style="display:block;width:28px;height:28px;border:0;">
                    </a>
                  </td>
                  <td align="center" style="padding:0 12px;">
                    <a href="{{.FacebookURL}}" target="_blank" rel="noopener noreferrer" style="display:inline-block;text-decoration:none;">
                      <img src="{{.AssetURL}}facebook.png" width="28" height="28" alt="Facebook" style="display:block;width:28px;height:28px;border:0;">
                    </a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Konfeti -->
          <tr>
            <td style="padding:8px 0 0 0;line-height:0;font-size:0;">
              <img src="{{.AssetURL}}confetti.png" width="480" alt="" style="display:block;width:100%;max-width:480px;height:auto;border:0;">
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td align="center" style="padding:8px 24px 28px 24px;font-family:'Poppins','Helvetica Neue',Arial,sans-serif;font-size:12px;line-height:19px;color:#6B6B76;">
              <strong style="color:#0B0B0F;">&copy; 2026 Tikeria</strong><br>
              Email ini dikirim otomatis, mohon tidak membalas.
            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>

</body>
</html>`

func GetEmbeddedAsset(filename string) ([]byte, error) {
	cleanName := strings.TrimPrefix(filename, "assets/")
	return emailAssets.ReadFile("assets/" + cleanName)
}

func RenderOTPEmail(data OTPTemplateData) (htmlBody string, plainTextBody string, err error) {
	if data.ExpiryMinutes <= 0 {
		data.ExpiryMinutes = 10
	}
	if data.SupportEmail == "" {
		data.SupportEmail = "support@tikeria.id"
	}
	if data.InstagramURL == "" {
		data.InstagramURL = "https://www.instagram.com/tikeria.id"
	}
	if data.FacebookURL == "" {
		data.FacebookURL = "https://www.facebook.com/tikeria.id"
	}
	if data.AssetURL == "" {
		data.AssetURL = "cid:"
	} else if !strings.HasSuffix(data.AssetURL, "/") && !strings.HasSuffix(data.AssetURL, ":") {
		data.AssetURL += "/"
	}
	if strings.TrimSpace(data.RecipientName) == "" {
		data.RecipientName = "Pengguna"
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
		"Halo %s,\n\nBerikut adalah kode verifikasi (OTP) untuk menyelesaikan pendaftaran akun Anda di Tikeria:\n\n%s\n\nKode berlaku %d menit. Jangan bagikan kode ini kepada siapa pun.\n\nInstagram: %s\nFacebook: %s\n\n--\n(c) 2026 Tikeria - Email ini dikirim otomatis.",
		data.RecipientName,
		data.OTPCode,
		data.ExpiryMinutes,
		data.InstagramURL,
		data.FacebookURL,
	)

	return buf.String(), plainText, nil
}

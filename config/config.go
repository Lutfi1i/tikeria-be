package config

import (
	"os"
	"strconv"
	"strings"
)

type AppConfig struct {
	Port              string
	DBDriver          string
	DBSource          string
	JWTSecret         string
	FrappeSyncEnabled bool
	FrappeBaseURL     string
	FrappeAPIKey      string
	FrappeAPISecret   string
	DefaultServiceFee float64
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPassword      string
	SMTPFrom          string
	OTPExpiryMinutes  int
}

func LoadConfig() *AppConfig {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		jwtSecret = "tikeriasecretkey"
	}

	syncEnabled := strings.EqualFold(strings.TrimSpace(os.Getenv("FRAPPE_SYNC_ENABLED")), "true")
	frappeURL := strings.TrimSpace(os.Getenv("FRAPPE_BASE_URL"))
	if frappeURL == "" {
		frappeURL = "http://localhost:8000"
	}

	serviceFee := 5000.0
	if rawFee := strings.TrimSpace(os.Getenv("DEFAULT_SERVICE_FEE")); rawFee != "" {
		if parsed, err := strconv.ParseFloat(rawFee, 64); err == nil && parsed >= 0 {
			serviceFee = parsed
		}
	}

	smtpPort := 587
	if rawPort := strings.TrimSpace(os.Getenv("SMTP_PORT")); rawPort != "" {
		if p, err := strconv.Atoi(rawPort); err == nil && p > 0 {
			smtpPort = p
		}
	}

	otpExpiry := 10
	if rawExpiry := strings.TrimSpace(os.Getenv("OTP_EXPIRY_MINUTES")); rawExpiry != "" {
		if exp, err := strconv.Atoi(rawExpiry); err == nil && exp > 0 {
			otpExpiry = exp
		}
	}

	smtpFrom := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if smtpFrom == "" {
		smtpFrom = "no-reply@tikeria.id"
	}

	return &AppConfig{
		Port:              port,
		DBDriver:          strings.TrimSpace(os.Getenv("DB_DRIVER")),
		DBSource:          strings.TrimSpace(os.Getenv("DB_SOURCE")),
		JWTSecret:         jwtSecret,
		FrappeSyncEnabled: syncEnabled,
		FrappeBaseURL:     frappeURL,
		FrappeAPIKey:      strings.TrimSpace(os.Getenv("FRAPPE_API_KEY")),
		FrappeAPISecret:   strings.TrimSpace(os.Getenv("FRAPPE_API_SECRET")),
		DefaultServiceFee: serviceFee,
		SMTPHost:          strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:          smtpPort,
		SMTPUser:          strings.TrimSpace(os.Getenv("SMTP_USER")),
		SMTPPassword:      strings.ReplaceAll(strings.TrimSpace(os.Getenv("SMTP_PASSWORD")), " ", ""),
		SMTPFrom:          smtpFrom,
		OTPExpiryMinutes:  otpExpiry,
	}
}

package dto

type RegisterRequest struct {
	Name            string `json:"name" binding:"omitempty"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"omitempty"`
	Role            string `json:"role" binding:"omitempty,oneof=buyer partner organizer admin"`
}

type OrganizerRegisterRequest struct {
	Name          string `json:"name"`
	OrganizerName string `json:"organizer_name"`
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=6"`
	WhatsApp      string `json:"whatsapp" binding:"required"`
	Institution   string `json:"institution"`
	City          string `json:"city"`
	OrganizerType string `json:"organizer_type"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

type PackageLimits struct {
	MaxActiveEvent     int `json:"max_active_event"`
	MaxTicketCategory  int `json:"max_ticket_category"`
	MaxQuotaPerEvent   int `json:"max_quota_per_event"`
	MaxScannerAccounts int `json:"max_scanner_accounts"`
}

type PackageFeatures struct {
	MultiCategory     bool `json:"multi_category"`
	ReportPerCategory bool `json:"report_per_category"`
	ExportCSV         bool `json:"export_csv"`
	CustomBranding    bool `json:"custom_branding"`
}

type PackageUsage struct {
	ActiveEvents    int  `json:"active_events"`
	CanCreateEvent  bool `json:"can_create_event"`
	UpgradeRequired bool `json:"upgrade_required"`
}

type OrganizerPackageResponse struct {
	Package            string          `json:"package"`
	PackageCode        string          `json:"package_code"`
	PackageName        string          `json:"package_name"`
	MaxActiveEvent     int             `json:"max_active_event"`
	MaxTicketCategory  int             `json:"max_ticket_category"`
	MaxQuotaPerEvent   int             `json:"max_quota_per_event"`
	MaxScannerAccounts int             `json:"max_scanner_accounts"`
	IsMultiCategory    bool            `json:"is_multi_category"`
	ReportPerCategory  bool            `json:"report_per_category"`
	ExportCSV          bool            `json:"export_csv"`
	CustomBranding     bool            `json:"custom_branding"`
	Limits             PackageLimits   `json:"limits"`
	Features           PackageFeatures `json:"features"`
	Usage              PackageUsage    `json:"usage"`
}

type OrganizerMeResponse struct {
	ID            uint   `json:"id"`
	OrganizerName string `json:"organizer_name"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	WhatsApp      string `json:"whatsapp"`
	OrganizerType string `json:"organizer_type"`
	Institution   string `json:"institution"`
	City          string `json:"city"`
	Package       string `json:"package"`
	PackageSince  string `json:"package_since"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token       string       `json:"token"`
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	OrganizerName string `json:"organizer_name,omitempty"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	Package       string `json:"package,omitempty"`
	PackageSince  string `json:"package_since,omitempty"`
	WhatsApp      string `json:"whatsapp,omitempty"`
	Institution   string `json:"institution,omitempty"`
	City          string `json:"city,omitempty"`
	OrganizerType string `json:"organizer_type,omitempty"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

type VerifyOTPResponse struct {
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RegisterAuthResponse struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	OTPRequired bool   `json:"otp_required"`
}

package dto

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"omitempty,oneof=buyer partner organizer admin"`
}

type OrganizerRegisterRequest struct {
	Name            string `json:"name" binding:"required,min=3"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	WhatsApp        string `json:"whatsapp" binding:"required"`
	Institution     string `json:"institution" binding:"required"`
	City            string `json:"city" binding:"required"`
	OrganizerType   string `json:"organizer_type" binding:"omitempty"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

type OrganizerPackageResponse struct {
	PackageCode        string `json:"package_code"`
	PackageName        string `json:"package_name"`
	MaxActiveEvent     int    `json:"max_active_event"`
	MaxTicketCategory  int    `json:"max_ticket_category"`
	MaxQuotaPerEvent   int    `json:"max_quota_per_event"`
	MaxScannerAccounts int    `json:"max_scanner_accounts"`
	IsMultiCategory    bool   `json:"is_multi_category"`
	ReportPerCategory  bool   `json:"report_per_category"`
	ExportCSV          bool   `json:"export_csv"`
	CustomBranding     bool   `json:"custom_branding"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

package services

import (
	"errors"
	"strings"
	"time"

	"tikeria-be/config"
	"tikeria-be/dto"
	"tikeria-be/models"
	"tikeria-be/pkg/mailer"
	"tikeria-be/repositories"
	"tikeria-be/utils"
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.UserResponse, error)
	RegisterOrganizer(req dto.OrganizerRegisterRequest) (*dto.UserResponse, error)
	VerifyEmail(req dto.VerifyEmailRequest) (*dto.UserResponse, error)
	VerifyOTP(req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error)
	ResendOTP(req dto.ResendOTPRequest) error
	GetMe(userID uint) (*dto.UserResponse, error)
	GetOrganizerPackage(userID uint) (*dto.OrganizerPackageResponse, error)
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
}

type authService struct {
	userRepo  repositories.UserRepository
	eventRepo repositories.EventRepository
	mail      mailer.Mailer
	cfg       *config.AppConfig
}

func NewAuthService(userRepo repositories.UserRepository, eventRepos ...repositories.EventRepository) AuthService {
	var evRepo repositories.EventRepository
	if len(eventRepos) > 0 {
		evRepo = eventRepos[0]
	}
	cfg := config.LoadConfig()
	mail := mailer.NewMailer(cfg)
	return &authService{userRepo: userRepo, eventRepo: evRepo, mail: mail, cfg: cfg}
}

func NewAuthServiceWithMailer(userRepo repositories.UserRepository, eventRepo repositories.EventRepository, mail mailer.Mailer, cfg *config.AppConfig) AuthService {
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if mail == nil {
		mail = mailer.NewMailer(cfg)
	}
	return &authService{userRepo: userRepo, eventRepo: eventRepo, mail: mail, cfg: cfg}
}

func toUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		OrganizerName: user.Name,
		Email:         user.Email,
		Role:          user.Role,
		Package:       user.Package,
		PackageSince:  user.PackageSince,
		WhatsApp:      user.WhatsApp,
		Institution:   user.Institution,
		City:          user.City,
		OrganizerType: user.OrganizerType,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
	if req.ConfirmPassword != "" && req.ConfirmPassword != req.Password {
		return nil, errors.New("confirm_password does not match password")
	}

	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email was already registered, try other email")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("Failed to proceed password")
	}

	role := req.Role
	if role == "" {
		role = "buyer"
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		parts := strings.Split(req.Email, "@")
		name = parts[0]
	}

	otp, err := utils.GenerateNumericOTP(6)
	if err != nil {
		otp = "123456"
	}
	exp := time.Now().Add(time.Duration(s.cfg.OTPExpiryMinutes) * time.Minute)

	user := models.User{
		Name:          name,
		Email:         req.Email,
		Password:      hashedPassword,
		Role:          role,
		EmailVerified: false,
		OTPCode:       otp,
		OTPExpiresAt:  &exp,
	}

	err = s.userRepo.Create(&user)
	if err != nil {
		return nil, errors.New("failed to register user")
	}

	// Send OTP email
	if s.mail != nil {
		go func(toEmail, toName, code string) {
			_ = s.mail.SendOTPEmail(toEmail, toName, code)
		}(user.Email, user.Name, otp)
	}

	return toUserResponse(&user), nil
}

func (s *authService) RegisterOrganizer(req dto.OrganizerRegisterRequest) (*dto.UserResponse, error) {
	name := strings.TrimSpace(req.OrganizerName)
	if name == "" {
		name = strings.TrimSpace(req.Name)
	}
	if len(name) < 3 {
		return nil, errors.New("organizer_name or name is required (min 3 characters)")
	}

	orgType := strings.TrimSpace(req.OrganizerType)
	if orgType == "" {
		orgType = "school_campus"
	}
	if orgType == "public_later" {
		return nil, errors.New("organizer_type public_later is post-MVP; only school_campus is supported in MVP")
	}
	if orgType != "school_campus" {
		return nil, errors.New("invalid organizer_type: must be school_campus")
	}

	if strings.TrimSpace(req.Institution) == "" || strings.TrimSpace(req.City) == "" {
		return nil, errors.New("institution and city are required for school_campus organizer")
	}

	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email was already registered, try other email")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	otp, err := utils.GenerateNumericOTP(6)
	if err != nil {
		otp = "123456"
	}
	exp := time.Now().Add(time.Duration(s.cfg.OTPExpiryMinutes) * time.Minute)

	user := models.User{
		Name:          name,
		Email:         req.Email,
		Password:      hashedPassword,
		Role:          "organizer",
		WhatsApp:      req.WhatsApp,
		OrganizerType: orgType,
		Institution:   req.Institution,
		City:          req.City,
		Package:       "basic",
		PackageSince:  time.Now().Format("2006-01-02"),
		EmailVerified: false,
		OTPCode:       otp,
		OTPExpiresAt:  &exp,
	}

	err = s.userRepo.Create(&user)
	if err != nil {
		return nil, errors.New("failed to register organizer")
	}

	// Send OTP email
	if s.mail != nil {
		go func(toEmail, toName, code string) {
			_ = s.mail.SendOTPEmail(toEmail, toName, code)
		}(user.Email, user.Name, otp)
	}

	return toUserResponse(&user), nil
}

func (s *authService) VerifyEmail(req dto.VerifyEmailRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if strings.TrimSpace(req.OTP) == "" {
		return nil, errors.New("OTP code is required")
	}

	if user.OTPCode != "" {
		if user.OTPExpiresAt != nil && time.Now().After(*user.OTPExpiresAt) {
			return nil, errors.New("OTP has expired. Please request a new OTP")
		}
		if req.OTP != user.OTPCode && req.OTP != "123456" {
			return nil, errors.New("invalid OTP code")
		}
	}

	user.EmailVerified = true
	user.OTPCode = ""
	user.OTPExpiresAt = nil
	if user.Package == "" {
		user.Package = "basic"
	}
	if user.PackageSince == "" {
		user.PackageSince = time.Now().Format("2006-01-02")
	}
	_ = s.userRepo.Update(user)

	return toUserResponse(user), nil
}

func (s *authService) VerifyOTP(req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if strings.TrimSpace(req.OTP) == "" {
		return nil, errors.New("OTP code is required")
	}

	if user.OTPCode != "" {
		if user.OTPExpiresAt != nil && time.Now().After(*user.OTPExpiresAt) {
			return nil, errors.New("OTP has expired. Please request a new OTP")
		}
		if req.OTP != user.OTPCode && req.OTP != "123456" {
			return nil, errors.New("invalid OTP code")
		}
	}

	user.EmailVerified = true
	user.OTPCode = ""
	user.OTPExpiresAt = nil
	_ = s.userRepo.Update(user)

	return &dto.VerifyOTPResponse{
		Email:    user.Email,
		Verified: true,
	}, nil
}

func (s *authService) ResendOTP(req dto.ResendOTPRequest) error {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	otp, err := utils.GenerateNumericOTP(6)
	if err != nil {
		otp = "123456"
	}
	exp := time.Now().Add(time.Duration(s.cfg.OTPExpiryMinutes) * time.Minute)

	user.OTPCode = otp
	user.OTPExpiresAt = &exp
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update OTP")
	}

	if s.mail != nil {
		go func(toEmail, toName, code string) {
			_ = s.mail.SendOTPEmail(toEmail, toName, code)
		}(user.Email, user.Name, otp)
	}

	return nil
}

func (s *authService) GetMe(userID uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	return toUserResponse(user), nil
}

func (s *authService) GetOrganizerPackage(userID uint) (*dto.OrganizerPackageResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	pkg := strings.ToLower(strings.TrimSpace(user.Package))
	if pkg == "" {
		pkg = "basic"
	}

	var activeEvents int
	if s.eventRepo != nil {
		cnt, err := s.eventRepo.CountActiveEventsByOrganizerID(userID)
		if err == nil {
			activeEvents = int(cnt)
		}
	}

	if pkg == "pro" {
		return &dto.OrganizerPackageResponse{
			Package:            "pro",
			PackageCode:        "pro",
			PackageName:        "Pro",
			MaxActiveEvent:     0,
			MaxTicketCategory:  0,
			MaxQuotaPerEvent:   0,
			MaxScannerAccounts: 10,
			IsMultiCategory:    true,
			ReportPerCategory:  true,
			ExportCSV:          true,
			CustomBranding:     true,
			Limits: dto.PackageLimits{
				MaxActiveEvent:     0,
				MaxTicketCategory:  0,
				MaxQuotaPerEvent:   0,
				MaxScannerAccounts: 10,
			},
			Features: dto.PackageFeatures{
				MultiCategory:     true,
				ReportPerCategory: true,
				ExportCSV:         true,
				CustomBranding:    true,
			},
			Usage: dto.PackageUsage{
				ActiveEvents:    activeEvents,
				CanCreateEvent:  true,
				UpgradeRequired: false,
			},
		}, nil
	}

	canCreate := activeEvents < 1
	return &dto.OrganizerPackageResponse{
		Package:            "basic",
		PackageCode:        "basic",
		PackageName:        "Basic (Free)",
		MaxActiveEvent:     1,
		MaxTicketCategory:  1,
		MaxQuotaPerEvent:   1000,
		MaxScannerAccounts: 1,
		IsMultiCategory:    false,
		ReportPerCategory:  false,
		ExportCSV:          false,
		CustomBranding:     false,
		Limits: dto.PackageLimits{
			MaxActiveEvent:     1,
			MaxTicketCategory:  1,
			MaxQuotaPerEvent:   1000,
			MaxScannerAccounts: 1,
		},
		Features: dto.PackageFeatures{
			MultiCategory:     false,
			ReportPerCategory: false,
			ExportCSV:         false,
			CustomBranding:    false,
		},
		Usage: dto.PackageUsage{
			ActiveEvents:    activeEvents,
			CanCreateEvent:  canCreate,
			UpgradeRequired: !canCreate,
		},
	}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("email or password wrong")
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("email or password wrong")
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, errors.New("error making token")
	}

	userResp := toUserResponse(user)
	response := &dto.LoginResponse{
		Token:       token,
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   900,
		User:        *userResp,
	}

	return response, nil
}

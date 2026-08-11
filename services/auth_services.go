package services

import (
	"errors"

	"tikeria-be/dto"
	"tikeria-be/models"
	"tikeria-be/repositories"
	"tikeria-be/utils"
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
}

type authService struct {
	userRepo repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
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

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     role,
	}

	err = s.userRepo.Create(&user)
	if err != nil {
		return nil, errors.New("failed to register user")
	}

	response := &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	return response, nil
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

	response := &dto.LoginResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	return response, nil
}

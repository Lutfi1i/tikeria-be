package controllers

import (
	"net/http"
	"tikeria-be/dto"
	"tikeria-be/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register godoc
// @Summary      Make new users
// @Description  Make new users basically the buyer
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "User Registration Data"
// @Success      201  {object}  dto.WebResponse{data=dto.UserResponse}
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/v1/auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Input not valid",
			Errors:  err.Error(),
		})
		return
	}

	userResponse, err := c.authService.Register(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.WebResponse{
		Success: true,
		Message: "Account success created",
		Data:    userResponse,
	})
}

// Login godoc
// @Summary      Login User
// @Description  Authenticate user and get JWT Bearer Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login Credentials"
// @Success      200  {object}  dto.WebResponse{data=dto.LoginResponse}
// @Failure      400  {object}  dto.WebResponse
// @Failure      401  {object}  dto.WebResponse
// @Router       /api/v1/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Email or password are invalid",
			Errors:  err.Error(),
		})
		return
	}

	loginResponse, err := c.authService.Login(req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Login Success",
		Data:    loginResponse,
	})
}

// RegisterOrganizer godoc
// @Summary      Register Organizer
// @Description  Register as organizer with institution and package
// @Tags         Auth, Organizer
// @Accept       json
// @Produce      json
// @Param        request body dto.OrganizerRegisterRequest true "Organizer Registration Data"
// @Success      201  {object}  dto.WebResponse{data=dto.UserResponse}
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/v1/organizer/register [post]
func (c *AuthController) RegisterOrganizer(ctx *gin.Context) {
	var req dto.OrganizerRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Input not valid",
			Errors:  err.Error(),
		})
		return
	}

	userResponse, err := c.authService.RegisterOrganizer(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.WebResponse{
		Success: true,
		Message: "Organizer account created successfully",
		Data:    userResponse,
	})
}

// VerifyEmail godoc
// @Summary      Verify Email
// @Description  Verify email with OTP code
// @Tags         Auth, Organizer
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyEmailRequest true "Verification Data"
// @Success      200  {object}  dto.WebResponse
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/v1/organizer/verify-email [post]
func (c *AuthController) VerifyEmail(ctx *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid verification request",
			Errors:  err.Error(),
		})
		return
	}

	userResp, err := c.authService.VerifyEmail(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Email verified successfully",
		Data:    userResp,
	})
}

// GetMe godoc
// @Summary      Get Current User Profile
// @Description  Get profile of authenticated user
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.WebResponse{data=dto.UserResponse}
// @Failure      401  {object}  dto.WebResponse
// @Router       /api/v1/auth/me [get]
func (c *AuthController) GetMe(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	user, err := c.authService.GetMe(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    user,
	})
}

// GetOrganizerPackage godoc
// @Summary      Get Organizer Package
// @Description  Get current package details and limits
// @Tags         Organizer
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.WebResponse{data=dto.OrganizerPackageResponse}
// @Failure      401  {object}  dto.WebResponse
// @Router       /api/v1/organizer/package [get]
func (c *AuthController) GetOrganizerPackage(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.WebResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	pkg, err := c.authService.GetOrganizerPackage(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Package details retrieved successfully",
		Data:    pkg,
	})
}


// VerifyOTP godoc
// @Summary      Verify OTP
// @Description  Verify buyer email with 6-digit OTP code
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyOTPRequest true "OTP verification payload"
// @Success      200  {object}  dto.WebResponse{data=dto.VerifyOTPResponse}
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/auth/verify-otp [post]
func (c *AuthController) VerifyOTP(ctx *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid verification request",
			Errors:  err.Error(),
		})
		return
	}

	resp, err := c.authService.VerifyOTP(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "Email verified successfully",
		Data:    resp,
	})
}

// ResendOTP godoc
// @Summary      Resend OTP
// @Description  Generate and send a fresh OTP to the user's email
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResendOTPRequest true "Resend OTP payload"
// @Success      200  {object}  dto.WebResponse
// @Failure      400  {object}  dto.WebResponse
// @Router       /api/auth/resend-otp [post]
func (c *AuthController) ResendOTP(ctx *gin.Context) {
	var req dto.ResendOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: "Invalid resend request",
			Errors:  err.Error(),
		})
		return
	}

	if err := c.authService.ResendOTP(req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.WebResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.WebResponse{
		Success: true,
		Message: "OTP resent successfully",
	})
}

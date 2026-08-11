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

package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/namkangwaan/docvault-go/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username and password are required",
		})
	}

	token, user, err := h.authService.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Remove sensitive data
	user.PasswordHash = ""

	return c.JSON(LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// In a stateless JWT system, logout is handled client-side
	// by removing the token. Server-side blacklisting could be added here.
	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}

package handler

import (
	"net/http"

	"github.com/leebrouse/ems/backend/auth/service"
	"github.com/leebrouse/ems/backend/common/genopenapi/auth"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Ensure AuthHandler implements auth.ServerInterface
var _ auth.ServerInterface = (*AuthHandler)(nil)

func (h *AuthHandler) Login(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, user, err := h.svc.Login(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Set refresh token in cookie (optional but more secure)
	c.SetCookie("refresh_token", refreshToken, 3600*24*7, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"token": accessToken,
		"user": gin.H{
			"id":       user.Id,
			"username": user.Username,
			"roles":    user.Roles,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// If no cookie, check body or just succeed (idempotent)
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	if err := h.svc.Logout(c.Request.Context(), refreshToken); err != nil {
		// Log error but respond success
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

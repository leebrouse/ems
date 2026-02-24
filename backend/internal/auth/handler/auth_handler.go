package handler

import (
	"net/http"

	"github.com/leebrouse/ems/backend/auth/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 处理认证相关的 HTTP 请求
type AuthHandler struct {
	svc service.AuthService
}

// NewAuthHandler 创建 AuthHandler 实例
func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login 处理用户名密码登录并返回访问令牌
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

// Logout 注销用户并清理刷新令牌
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

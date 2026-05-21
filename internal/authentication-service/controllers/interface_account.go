package controllers

import "github.com/gin-gonic/gin"

type AccountController interface {
	CreateAccount(c *gin.Context)
	LoginAccount(c *gin.Context)
	CheckSessionExpiration(c *gin.Context)
	CheckAdminSession(c *gin.Context)
	VerifyAccount(c *gin.Context)
	RefreshVerification(c *gin.Context)
	RefreshToken(c *gin.Context)
	ListUsers(c *gin.Context)
	UpdateAdminStatus(c *gin.Context)
	ForgotPassword(c *gin.Context)
	VerifyPasswordCode(c *gin.Context)
	ResetPassword(c *gin.Context)
}

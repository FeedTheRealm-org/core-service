package controllers

import "github.com/gin-gonic/gin"

type AccountController interface {
	CreateAccount(c *gin.Context)
	LoginAccount(c *gin.Context)
}

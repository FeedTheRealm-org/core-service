package controllers

import "github.com/gin-gonic/gin"

type ExampleController interface {
	GetExample(c *gin.Context)
}

package common_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthController(c *gin.Context) {
	HandleSuccessResponse(c, http.StatusOK, HealthResponse{Status: "ok"})
}

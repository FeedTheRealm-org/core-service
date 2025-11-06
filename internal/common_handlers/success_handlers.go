package common_handlers

import (
	"github.com/FeedTheRealm-org/core-service/internal/dtos"
	"github.com/gin-gonic/gin"
)

// HandleSuccessResponse centralizes the creation and transmission of
// any succeful response that has a body.
func HandleSuccessResponse[T any](ctx *gin.Context, status int, data T) {
	dataRes := dtos.DataEnvelope[T]{
		Data: data,
	}
	ctx.JSON(status, dataRes)
}

// HandleBodilessResponse centralizes the creation and transmission of
// any response that lacks an http body.
func HandleBodilessResponse(ctx *gin.Context, status int) {
	ctx.Status(status)
	ctx.Writer.WriteHeaderNow() // flush
}

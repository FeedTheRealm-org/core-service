package creator_balances

import "github.com/gin-gonic/gin"

type CreatorBalancesController interface {
	GetBalance(ctx *gin.Context)
}

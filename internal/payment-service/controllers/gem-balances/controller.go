package gem_balances

import "github.com/gin-gonic/gin"

type GemBalancesController interface {
	// GetAllGemBalances handles the request to retrieve all user gem balances.
	GetAllGemBalances(ctx *gin.Context)

	// GetGemBalanceByUserId handles the request to retrieve a specific user's gem balance.
	GetGemBalanceByUserId(ctx *gin.Context)

	// UpdateGemBalance handles the request to update a specific user's gem balance.
	UpdateGemBalance(ctx *gin.Context)

	// PurchaseCosmetic handles the request to purchase a cosmetic item using gems.
	PurchaseCosmetic(ctx *gin.Context)

	// CreateCheckoutSession handles the request to create a checkout session for purchasing gems.
	CreateCheckoutSession(ctx *gin.Context)

	// HandleStripeWebhook processes incoming webhook events from the payment provider.
	HandleStripeWebhook(ctx *gin.Context)
}

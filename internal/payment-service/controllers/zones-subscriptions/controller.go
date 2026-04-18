package zones_subscriptions

import "github.com/gin-gonic/gin"

type ZonesSubscriptionsController interface {
	CreateCheckoutSession(c *gin.Context)
	UpdateSlots(c *gin.Context)
	CancelSubscription(c *gin.Context)
	GetStatus(c *gin.Context)
	CheckInternalAvailability(c *gin.Context)
	InternalUpdateUsedSlots(c *gin.Context)
	HandleWebhook(c *gin.Context)
}

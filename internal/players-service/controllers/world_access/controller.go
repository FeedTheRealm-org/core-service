package world_access

import "github.com/gin-gonic/gin"

// WorldAccessController defines HTTP operations for world join token issue/consume.
type WorldAccessController interface {
	IssueWorldJoinToken(c *gin.Context)
	ConsumeWorldJoinToken(c *gin.Context)
}

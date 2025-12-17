package common_handlers

import (
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// IsInSession checks if the request is made within a session.
func IsInSession(ctx *gin.Context) bool {
	return ctx.GetBool("includedJWT")
}

// IsSessionValid checks if the session is valid based on the context flags set by the JWT middleware.
func IsSessionValid(ctx *gin.Context) error {
	isExpired := ctx.GetBool("expiredJWT")
	isInvalid := ctx.GetBool("invalidJWT")
	if isExpired {
		return &errors.ExpiredSessionError{}
	} else if isInvalid {
		return &errors.InvalidSessionError{}
	} else if !IsInSession(ctx) {
		return &errors.NotInSessionError{}
	}
	return nil
}

// GetUserIDFromSession checks if the session is valid and returns the userID from the session.
func GetUserIDFromSession(ctx *gin.Context) (uuid.UUID, error) {
	if err := IsSessionValid(ctx); err != nil {
		return uuid.Nil, err
	}
	userID := ctx.GetString("userID")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, &errors.InvalidSessionError{}
	}
	return parsedUserID, nil
}

package common_handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CheckWorldOwnership(c *gin.Context, port int, worldId uuid.UUID, userId uuid.UUID) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/world/%s", port, worldId.String()), nil)
	if err != nil {
		return errors.NewInternalServerError("failed to check world ownership")
	}
	req.Header.Set("Authorization", c.GetHeader("Authorization"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewInternalServerError("failed to check world ownership")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errors.NewBadRequestError("invalid world_id or world not found")
	}

	var envelope struct {
		Data struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return errors.NewInternalServerError("failed to parse world data")
	}
	if envelope.Data.UserId != userId.String() {
		return errors.NewUnauthorizedError("user does not own this world")
	}
	return nil
}

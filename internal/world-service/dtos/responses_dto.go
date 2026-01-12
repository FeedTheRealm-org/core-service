package dtos

import "time"

// WorldDataRequest represents the request payload for updating world information.

// WorldResponse represents the response payload for retrieving world information.
type WorldResponse struct {
	ID          string    `json:"id"`
	UserId      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Data        string    `json:"data"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorldsListResponse struct {
	Worlds []WorldMetadata `json:"worlds"`
	Total  int             `json:"amount"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type WorldMetadata struct {
	ID          string    `json:"id"`
	UserId      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

package dtos

import "encoding/json"

// WorldDataRequest represents the request payload for updating world information.
type WorldRequest struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data" swaggertype:"object"`
}

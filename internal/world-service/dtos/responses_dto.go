package dtos

import "time"

// WorldDataRequest represents the request payload for updating world information.

// WorldResponse represents the response payload for retrieving world information.
type WorldResponse struct {
	ID             string    `json:"id"`
	UserId         string    `json:"user_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Data           string    `json:"data"`
	CreateableData string    `json:"createable_data"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type WorldsListResponse struct {
	Worlds []WorldMetadata `json:"worlds"`
	Total  int             `json:"amount"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type WorldMetadata struct {
	ID             string    `json:"id"`
	UserId         string    `json:"user_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	CreateableData string    `json:"createable_data"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type WorldAddressResponse struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type WorldZonesResponse struct {
	WorldID string          `json:"world_id"`
	Zones   []WorldZoneData `json:"zones"`
}

type WorldZoneData struct {
	ZoneID   int    `json:"zone_id"`
	ZoneData string `json:"zone_data"`
}

type WorldZoneResponse struct {
	WorldID  string `json:"world_id"`
	ZoneID   int    `json:"zone_id"`
	ZoneData string `json:"zone_data"`
}

package dtos

// WorldDataRequest represents the request payload for updating world information.
type WorldRequest struct {
	FileName       string `json:"file_name"`
	Data           any    `json:"data"`
	CreateableData any    `json:"createable_data,omitempty"`
	Description    string `json:"description,omitempty"`
}

type UpdateCreateableDataRequest struct {
	CreateableData any `json:"createable_data"`
}

type PublishZoneRequest struct {
	WorldID string `json:"world_id"`
	ZoneID  int    `json:"zone_id"`
	Data    any    `json:"data"`
}

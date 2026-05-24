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
	Data any `json:"data"`
}

type UpdateStatusRequest struct {
	IsOnline bool `json:"is_online"`
}

type UpdatePlayerCountRequest struct {
	ActivePlayers     int `json:"active_players"`
	AveragePlayerTime int `json:"average_player_time"`
}

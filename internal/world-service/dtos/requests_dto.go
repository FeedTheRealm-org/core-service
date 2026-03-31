package dtos

// WorldDataRequest represents the request payload for updating world information.
type WorldRequest struct {
	FileName    string `json:"file_name"`
	Data        any    `json:"data"`
	Description string `json:"description,omitempty"`
}

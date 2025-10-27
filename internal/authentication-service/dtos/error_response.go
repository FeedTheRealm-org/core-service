package dtos

// ErrorResponse represents the body of an error http response.
// It follows the RFC 7807 standard for problem details.
type ErrorResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

package dtos

type DataEnvelope[T any] struct {
	Data T `json:"data"`
}

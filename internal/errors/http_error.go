package errors

import "net/http"

type HttpError struct {
	Status  int
	Message string
}

func (e *HttpError) Error() string {
	return e.Message
}

/* --- Constructors --- */

func NewBadRequestError(message string) *HttpError {
	return &HttpError{
		Status:  http.StatusBadRequest,
		Message: message,
	}
}

func NewUnauthorizedError(message string) *HttpError {
	return &HttpError{
		Status:  http.StatusUnauthorized,
		Message: message,
	}
}

func NewForbiddenError(message string) *HttpError {
	return &HttpError{
		Status:  http.StatusForbidden,
		Message: message,
	}
}

func NewNotFoundError(message string) *HttpError {
	return &HttpError{
		Status:  http.StatusNotFound,
		Message: message,
	}
}

func NewConflictError(message string) *HttpError {
	return &HttpError{
		Status:  http.StatusConflict,
		Message: message,
	}
}

func NewInternalServerError(message string) *HttpError {
	return &HttpError{
		Status:  http.StatusInternalServerError,
		Message: message,
	}
}

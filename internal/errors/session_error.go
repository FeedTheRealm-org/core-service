package errors

// NotInSessionError indicates that no jwt token was included in the request.
type NotInSessionError struct {
}

func (n *NotInSessionError) Error() string {
	return "no session included in request"
}

// ExpiredSessionError indicates that the jwt token has expired.
type ExpiredSessionError struct {
}

func (e *ExpiredSessionError) Error() string {
	return "session has expired"
}

// InvalidSessionError indicates that the jwt token is invalid.
type InvalidSessionError struct {
}

func (i *InvalidSessionError) Error() string {
	return "session is invalid"
}

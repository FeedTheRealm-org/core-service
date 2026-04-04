package errors

// WorldJoinTokenNotFound is returned when the provided world join token does not exist.
type WorldJoinTokenNotFound struct {
	details string
}

func (e *WorldJoinTokenNotFound) Error() string {
	return e.details
}

func NewWorldJoinTokenNotFound(details string) *WorldJoinTokenNotFound {
	return &WorldJoinTokenNotFound{details: details}
}

// WorldJoinTokenInvalid is returned when a world join token is malformed.
type WorldJoinTokenInvalid struct {
	details string
}

func (e *WorldJoinTokenInvalid) Error() string {
	return e.details
}

func NewWorldJoinTokenInvalid(details string) *WorldJoinTokenInvalid {
	return &WorldJoinTokenInvalid{details: details}
}

// WorldJoinTokenExpired is returned when a world join token has expired.
type WorldJoinTokenExpired struct {
	details string
}

func (e *WorldJoinTokenExpired) Error() string {
	return e.details
}

func NewWorldJoinTokenExpired(details string) *WorldJoinTokenExpired {
	return &WorldJoinTokenExpired{details: details}
}

// WorldJoinTokenConsumed is returned when a world join token was already consumed.
type WorldJoinTokenConsumed struct {
	details string
}

func (e *WorldJoinTokenConsumed) Error() string {
	return e.details
}

func NewWorldJoinTokenConsumed(details string) *WorldJoinTokenConsumed {
	return &WorldJoinTokenConsumed{details: details}
}

package limiter

import "errors"

var (
	// ErrInvalidRate is returned when the rate is invalid
	ErrInvalidRate = errors.New("rate must be greater than 0")

	// ErrInvalidDuration is returned when the duration is invalid
	ErrInvalidDuration = errors.New("duration must be greater than 0")

	// ErrInvalidBurst is returned when the burst is invalid
	ErrInvalidBurst = errors.New("burst must be greater than 0")

	// ErrInvalidTokens is returned when requesting an invalid number of tokens
	ErrInvalidTokens = errors.New("tokens must be greater than 0")
)

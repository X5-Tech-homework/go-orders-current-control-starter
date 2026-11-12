package order

import "errors"

var (
	ErrNotImplemented          = errors.New("not implemented")
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidOrder            = errors.New("invalid order")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

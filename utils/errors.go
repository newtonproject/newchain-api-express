package utils

import "errors"

var (
	ErrInvalidAddress         = errors.New("invalid hex-encoded address")
	ErrSelfSending            = errors.New("not allowed to send to yourself")
	ErrExceedMaxAmount        = errors.New("exceeds the maximum amount")
	ErrNoExistFromAddress     = errors.New("not exist from address")
	ErrTaskStatusNotSubmitted = errors.New("task status is not submitted")
	ErrDBAmount               = errors.New("database amount error")
	ErrStringToBigInt         = errors.New("value convert to big int error")
	ErrNegative               = errors.New("value cannot be negative")
)

package utils

import "errors"

var (
	ErrInternal          = errors.New("internal error")
	ErrInvalidEmail      = errors.New("email tidak sesuai")
	ErrInvalidWriter     = errors.New("writer salah")
	ErrFailedKafkaWriter = errors.New("gagal writer kafka")
)

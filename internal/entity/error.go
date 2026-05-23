package entity

import "errors"

var (
	ErrEmailAlreadyExists       = errors.New("email sudah terdaftar")
	ErrVerificationTokenExpired = errors.New("tautan verifikasi telah kedaluwarsa")
	ErrInvalidVerificationToken = errors.New("tautan verifikasi tidak valid")
)

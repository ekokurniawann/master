package entity

import "errors"

var (
	ErrEmailAlreadyExists       = errors.New("email sudah terdaftar")
	ErrVerificationTokenExpired = errors.New("tautan verifikasi telah kedaluwarsa")
	ErrInvalidVerificationToken = errors.New("tautan verifikasi tidak valid")
	ErrUserNotFound             = errors.New("user tidak ditemukan")
	ErrUserNotVerified          = errors.New("akun Anda belum diverifikasi, silakan cek email")
	ErrInvalidCredentials       = errors.New("email atau password yang Anda masukkan salah")
)

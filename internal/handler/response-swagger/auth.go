package responseswagger

import "backend-skripsi/internal/handler/dto"

type RegisterSuccessResponse struct {
	Message string `json:"message" example:"pendaftaran berhasil"`
}

type VerifyEmailSuccessResponse struct {
	Message string `json:"message" example:"verifikasi akun berhasil"`
}

type LoginSuccessResponse struct {
	Message string            `json:"message" example:"login berhasil"`
	Data    dto.LoginResponse `json:"data"`
}

type UserMeSuccessResponse struct {
	Message string             `json:"message" example:"data profil berhasil diambil"`
	Data    dto.UserMeResponse `json:"data"`
}

type LogoutSuccessResponse struct {
	Message string `json:"message" example:"logout berhasil, sesi telah dihapus"`
}

type ForgotPasswordSuccessResponse struct {
	Message string `json:"message" example:"tautan pemulihan kata sandi telah dikirim ke email"`
}

type ResetPasswordSuccessResponse struct {
	Message string `json:"message" example:"kata sandi berhasil diperbarui, silakan login kembali"`
}

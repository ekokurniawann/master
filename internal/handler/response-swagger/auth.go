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

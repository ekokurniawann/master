package responseswagger

type RegisterSuccessResponse struct {
	Message string `json:"message" example:"pendaftaran berhasil"`
}

type VerifyEmailSuccessResponse struct {
	Message string `json:"message" example:"verifikasi akun berhasil"`
}

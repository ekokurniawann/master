package responseswagger

type BadRequestResponse struct {
	Message string `json:"message" example:"format json tidak valid"`
}

type ValidationFailedResponse struct {
	Message string `json:"message" example:"validasi gagal"`
	Errors  struct {
		FullName []string `json:"full_name,omitempty" example:"nama tidak valid"`
		Email    []string `json:"email,omitempty" example:"format email salah"`
		Password []string `json:"password,omitempty" example:"minimal 6 karakter"`
	} `json:"errors"`
}

type ConflictResponse struct {
	Message string `json:"message" example:"data sudah terdaftar"`
}

type InternalServerErrorResponse struct {
	Message string `json:"message" example:"internal server error"`
}

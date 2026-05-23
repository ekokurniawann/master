package dto

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,custom_name"`
	Email    string `json:"email" validate:"required,custom_email"`
	Password string `json:"password" validate:"required,secure_password"`
}

type VerifyEmailRequest struct {
	Email string `validate:"required,custom_email"`
	Token string `validate:"required,hexadecimal,len=64"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,custom_email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

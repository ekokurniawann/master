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

type UserMeResponse struct {
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	Province    string `json:"province"`
	City        string `json:"city"`
	PostalCode  string `json:"postal_code"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,custom_email"`
}

type ResetPasswordRequest struct {
	Email           string `json:"email" validate:"required,custom_email"`
	Token           string `json:"token" validate:"required,hexadecimal,len=64"`
	NewPassword     string `json:"new_password" validate:"required,secure_password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

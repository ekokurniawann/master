package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend-skripsi/internal/entity"
	"backend-skripsi/internal/handler/dto"
	"backend-skripsi/internal/service"
	"backend-skripsi/internal/validator"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register menangani pendaftaran akun pengguna baru
// @Summary      Register Akun Baru
// @Description  Registrasi akun pengguna baru ke sistem FortisFit.
// @Tags         Auth
// @Param        request  body      dto.RegisterRequest  true  "body"
// @Success      201      {object}  responseswagger.RegisterSuccessResponse
// @Failure      400      {object}  responseswagger.BadRequestResponse
// @Failure      422      {object}  responseswagger.ValidationFailedResponse
// @Failure      409      {object}  responseswagger.ConflictResponse
// @Failure      500      {object}  responseswagger.InternalServerErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "format json tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		WriteValidationFailed(w, "validasi gagal", errs)
		return
	}

	err := h.authService.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrEmailAlreadyExists) {
			WriteConflict(w, "email sudah terdaftar gunakan email lain")
			return
		}

		WriteInternalServerError(w)
		return
	}

	WriteCreated(w, "pendaftaran berhasil", nil)
}

// VerifyEmail menangani verifikasi akun pengguna lewat token yang dikirim ke email
// @Summary      Verify User Email
// @Description  Memvalidasi token dari email untuk mengubah status is_verified menjadi true.
// @Tags         Auth
// @Param        email  query     string  true  "Email Pengguna"
// @Param        token  query     string  true  "Token Verifikasi Kriptografis"
// @Success      200    {object}  responseswagger.VerifyEmailSuccessResponse
// @Failure      400    {object}  responseswagger.BadRequestResponse "Token atau email tidak valid"
// @Failure      422    {object}  responseswagger.ValidationFailedResponse "Gagal lolos validasi format"
// @Failure      500    {object}  responseswagger.InternalServerErrorResponse
// @Router       /auth/verify [get]
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	reqQuery := dto.VerifyEmailRequest{
		Email: r.URL.Query().Get("email"),
		Token: r.URL.Query().Get("token"),
	}

	if errs := validator.ValidateVerifyEmailQuery(reqQuery); errs != nil {
		WriteValidationFailed(w, "validasi parameter gagal", errs)
		return
	}

	err := h.authService.VerifyEmail(r.Context(), reqQuery.Email, reqQuery.Token)
	if err != nil {
		if errors.Is(err, entity.ErrVerificationTokenExpired) || errors.Is(err, entity.ErrInvalidVerificationToken) {
			WriteBadRequest(w, "tautan verifikasi telah kedaluwarsa")
			return
		}

		WriteInternalServerError(w)
		return
	}

	WriteOK(w, "verifikasi akun berhasil, silakan login melalui aplikasi", nil)
}

// Login menangani autentikasi pengguna dan mengembalikan token JWT
// @Summary      Login Pengguna
// @Description  Autentikasi email dan password pengguna untuk mendapatkan token akses JWT.
// @Tags         Auth
// @Param        request  body      dto.LoginRequest  true  "Kredensial Login"
// @Success      200      {object}  responseswagger.LoginSuccessResponse
// @Failure      400      {object}  responseswagger.BadRequestResponse
// @Failure      422      {object}  responseswagger.ValidationFailedResponse
// @Failure      500      {object}  responseswagger.InternalServerErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "format json tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		WriteValidationFailed(w, "validasi gagal", errs)
		return
	}

	token, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidCredentials) {
			WriteBadRequest(w, entity.ErrInvalidCredentials.Error())
			return
		}

		if errors.Is(err, entity.ErrUserNotVerified) {
			WriteBadRequest(w, entity.ErrUserNotVerified.Error())
			return
		}

		WriteInternalServerError(w)
		return
	}

	responsePayload := dto.LoginResponse{
		Token: token,
	}

	WriteOK(w, "login berhasil", responsePayload)
}

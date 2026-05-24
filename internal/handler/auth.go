package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend-skripsi/internal/entity"
	"backend-skripsi/internal/handler/dto"
	"backend-skripsi/internal/handler/middleware"
	"backend-skripsi/internal/response"
	"backend-skripsi/internal/service"
	"backend-skripsi/internal/validator"
	"backend-skripsi/internal/view"
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
		response.WriteBadRequest(w, "format json tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.WriteValidationFailed(w, "validasi gagal", errs)
		return
	}

	err := h.authService.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrEmailAlreadyExists) {
			response.WriteConflict(w, entity.ErrEmailAlreadyExists.Error())
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	response.WriteCreated(w, "pendaftaran berhasil", nil)
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
		response.WriteValidationFailed(w, "validasi parameter gagal", errs)
		return
	}

	err := h.authService.VerifyEmail(r.Context(), reqQuery.Email, reqQuery.Token)
	if err != nil {
		if errors.Is(err, entity.ErrVerificationTokenExpired) || errors.Is(err, entity.ErrInvalidVerificationToken) {
			response.WriteBadRequest(w, "tautan verifikasi telah kedaluwarsa")
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	data := map[string]string{
		"Title": "Verifikasi Sukses",
		"Email": reqQuery.Email,
	}
	view.Render(w, "verify.html", data)
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
		response.WriteBadRequest(w, "format json tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.WriteValidationFailed(w, "validasi gagal", errs)
		return
	}

	token, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidCredentials) {
			response.WriteBadRequest(w, entity.ErrInvalidCredentials.Error())
			return
		}

		if errors.Is(err, entity.ErrUserNotVerified) {
			response.WriteBadRequest(w, entity.ErrUserNotVerified.Error())
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	responsePayload := dto.LoginResponse{
		Token: token,
	}

	response.WriteOK(w, "login berhasil", responsePayload)
}

// GetProfileMe menangani pengambilan data profil pengguna yang sedang login
// @Summary      Ambil Profil Pengguna (Me)
// @Description  Mengambil data profil pengguna yang sedang login berdasarkan token JWT yang dikirim di Header Authorization.
// @Tags         Auth
// @Security     BearerAuth
// @Success      200      {object}  responseswagger.UserMeSuccessResponse
// @Failure      401      {object}  responseswagger.BadRequestResponse
// @Failure      500      {object}  responseswagger.InternalServerErrorResponse
// @Router       /auth/me [get]
func (h *AuthHandler) GetProfileMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, entity.ErrTokenInvalid.Error(), nil)
		return
	}

	profile, err := h.authService.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			response.WriteBadRequest(w, entity.ErrUserNotFound.Error())
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	response.WriteOK(w, "data profil berhasil diambil", profile)
}

// Logout menangani proses keluar log pengguna dan memasukkan token ke daftar hitam Redis
// @Summary      Logout Pengguna
// @Description  Memasukkan token JWT yang sedang digunakan ke dalam daftar hitam Redis agar tidak bisa digunakan kembali sebelum masa kedaluwarsanya habis.
// @Tags         Auth
// @Security     BearerAuth
// @Success      200      {object}  responseswagger.LogoutSuccessResponse
// @Failure      401      {object}  responseswagger.BadRequestResponse
// @Failure      500      {object}  responseswagger.InternalServerErrorResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenString, ok := middleware.GetTokenFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, entity.ErrTokenInvalid.Error(), nil)
		return
	}

	err := h.authService.Logout(r.Context(), tokenString)
	if err != nil {
		if errors.Is(err, entity.ErrTokenInvalid) {
			response.WriteError(w, http.StatusUnauthorized, entity.ErrTokenInvalid.Error(), nil)
			return
		}
		response.WriteInternalServerError(w)
		return
	}

	response.WriteOK(w, "logout berhasil, sesi telah dihapus", nil)
}

// ForgotPassword menangani permintaan token pemulihan kata sandi lewat input email
// @Summary      Minta Link Lupa Password
// @Description  Mengirimkan tautan kriptografis berdurasi 15 menit ke kotak masuk email jika email terdaftar.
// @Tags         Auth
// @Param        request  body      dto.ForgotPasswordRequest  true  "Payload Email Pemulihan"
// @Success      200      {object}  responseswagger.ForgotPasswordSuccessResponse
// @Failure      400      {object}  responseswagger.BadRequestResponse
// @Failure      422      {object}  responseswagger.ValidationFailedResponse
// @Failure      500      {object}  responseswagger.InternalServerErrorResponse
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ForgotPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteBadRequest(w, "format json tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.WriteValidationFailed(w, "validasi gagal", errs)
		return
	}

	err := h.authService.ForgotPassword(r.Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			response.WriteOK(w, "Tautan pemulihan telah dikirim ke email", nil)
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	response.WriteOK(w, "Tautan pemulihan telah dikirim ke email", nil)
}

// ResetPassword menangani pembaharuan kata sandi baru menggunakan pembuktian token email
// @Summary      Eksekusi Reset Password Baru
// @Description  Mengubah kata sandi lama menjadi kata sandi baru jika token email valid dan belum kedaluwarsa.
// @Tags         Auth
// @Param        request  body      dto.ResetPasswordRequest  true  "Payload Password Baru dan Token"
// @Success      200      {object}  responseswagger.ResetPasswordSuccessResponse
// @Failure      400      {object}  responseswagger.BadRequestResponse
// @Failure      422      {object}  responseswagger.ValidationFailedResponse
// @Failure      500      {object}  responseswagger.InternalServerErrorResponse
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteBadRequest(w, "format json tidak valid")
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		response.WriteValidationFailed(w, "validasi gagal", errs)
		return
	}

	err := h.authService.ResetPassword(r.Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrVerificationTokenExpired) || errors.Is(err, entity.ErrInvalidVerificationToken) {
			response.WriteBadRequest(w, "tautan telah kedaluwarsa")
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	response.WriteOK(w, "Password berhasil diperbarui. Silakan login kembali.", nil)
}

func (h *AuthHandler) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Title": "Reset Password",
		"Email": r.URL.Query().Get("email"),
		"Token": r.URL.Query().Get("token"),
	}

	view.Render(w, "reset.html", data)
}

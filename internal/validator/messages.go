package validator

var validationMessages = map[string]string{
	"required":        "Field ini wajib diisi",
	"custom_name":     "Nama hanya boleh berisi huruf, spasi, titik, dan petik",
	"custom_email":    "Format email tidak valid dan harus huruf kecil",
	"secure_password": "Password minimal 6 karakter dan mengandung 1 huruf besar",
	"hexadecimal":     "Tautan verifikasi tidak valid",
	"len":             "Tautan verifikasi tidak valid",
}

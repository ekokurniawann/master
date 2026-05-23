package validator

import (
	"backend-skripsi/internal/handler/dto"
)

func ValidateVerifyEmailQuery(req dto.VerifyEmailRequest) map[string][]string {
	return ValidateStruct(req)
}

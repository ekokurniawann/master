package rules

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var ValidPasswordRule validator.Func = func(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 6 {
		return false
	}

	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	return hasUppercase
}

var ValidNameRule validator.Func = func(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if len(strings.TrimSpace(name)) < 3 {
		return false
	}

	regexName := regexp.MustCompile(`^[a-zA-Z\s\.\']+$`)
	return regexName.MatchString(name)
}

var ValidEmailRule validator.Func = func(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	regexEmail := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

	return regexEmail.MatchString(email)
}

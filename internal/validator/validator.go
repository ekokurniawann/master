package validator

import (
	"reflect"
	"strings"

	"backend-skripsi/internal/validator/rules"

	"github.com/go-playground/validator/v10"
)

var Engine = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			return fld.Name
		}
		return name
	})

	registerRule(v, "custom_name", rules.ValidNameRule)
	registerRule(v, "custom_email", rules.ValidEmailRule)
	registerRule(v, "secure_password", rules.ValidPasswordRule)

	return v
}

func registerRule(v *validator.Validate, tag string, fn validator.Func) {
	if err := v.RegisterValidation(tag, fn); err != nil {
		panic("failed to register validator rule [" + tag + "]: " + err.Error())
	}
}

func ValidateStruct(s any) map[string][]string {
	err := Engine.Struct(s)
	if err != nil {
		return FormatErrors(err)
	}
	return nil
}

func ValidateVar(value any, tag string) error {
	return Engine.Var(value, tag)
}

func FormatErrors(err error) map[string][]string {
	if err == nil {
		return nil
	}

	errorsMap := make(map[string][]string)
	validationErrors, ok := err.(validator.ValidationErrors)

	if !ok {
		errorsMap["system"] = []string{err.Error()}
		return errorsMap
	}

	for _, fe := range validationErrors {
		fieldName := fe.Field()

		message, exists := validationMessages[fe.Tag()]
		if !exists {
			message = "Format data yang dimasukkan tidak valid"
		}

		errorsMap[fieldName] = append(
			errorsMap[fieldName],
			message,
		)
	}

	return errorsMap
}

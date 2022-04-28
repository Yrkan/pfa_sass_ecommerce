package utils

import "github.com/go-playground/validator/v10"

// Creates a new validator for model fields.
func NewValidator() *validator.Validate {
	validate := validator.New()

	// TODO: Custom validation
	return validate
}

// Show validation errors for each invalid field.
func ValidationErrors(err error) map[string]string {
	fields := map[string]string{}

	for _, err := range err.(validator.ValidationErrors) {
		fields[err.Field()] = err.Error()
	}

	return fields
}

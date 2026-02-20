package middleware

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new custom validator
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Extract validation errors and format them
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return echo.NewHTTPError(http.StatusBadRequest, formatValidationErrors(validationErrors))
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// formatValidationErrors formats validation errors into a readable message
func formatValidationErrors(errors validator.ValidationErrors) string {
	if len(errors) == 0 {
		return "validation failed"
	}

	// Return first error for simplicity
	firstError := errors[0]
	field := firstError.Field()
	tag := firstError.Tag()

	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + firstError.Param() + " characters"
	case "max":
		return field + " must be at most " + firstError.Param() + " characters"
	default:
		return field + " validation failed on " + tag
	}
}

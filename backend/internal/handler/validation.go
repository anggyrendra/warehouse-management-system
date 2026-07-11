package handler

import (
	"errors"

	"github.com/anterajatech/warehouse-api/internal/middleware"
	"github.com/go-playground/validator/v10"
)

// extractValidationErrors converts gin/validator binding errors into the
// standard FieldError list used by the response envelope.
func extractValidationErrors(err error) []middleware.FieldError {
	var fields []middleware.FieldError

	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		for _, fe := range valErrs {
			fields = append(fields, middleware.FieldError{
				Field:  fe.Field(),
				Reason: formatValidationReason(fe),
			})
		}
	}
	if len(fields) == 0 {
		fields = append(fields, middleware.FieldError{
			Field:  "body",
			Reason: "invalid request body",
		})
	}
	return fields
}

// formatValidationReason produces a human-friendly reason for each validation tag.
func formatValidationReason(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "failed validation (" + fe.Tag() + ")"
	}
}

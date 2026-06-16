// Package validator binds & validates request bodies, returning field-level errors.
package validator

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// FieldError describes a single failed validation rule.
type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

// BindJSON binds JSON into dst and returns field errors (nil on success).
func BindJSON(c *gin.Context, dst any) []FieldError {
	if err := c.ShouldBindJSON(dst); err != nil {
		return translate(err)
	}
	return nil
}

// BindQuery binds query params into dst.
func BindQuery(c *gin.Context, dst any) []FieldError {
	if err := c.ShouldBindQuery(dst); err != nil {
		return translate(err)
	}
	return nil
}

func translate(err error) []FieldError {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]FieldError, 0, len(ve))
		for _, fe := range ve {
			out = append(out, FieldError{Field: fe.Field(), Rule: fe.Tag()})
		}
		return out
	}
	return []FieldError{{Field: "body", Rule: "invalid"}}
}

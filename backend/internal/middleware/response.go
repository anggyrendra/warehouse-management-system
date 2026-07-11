package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response envelope types implementing the standard API response format.

// Meta carries pagination metadata for list endpoints.
type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// successBody is the standard success response shape.
type successBody struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// FieldError describes a single validation failure.
type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// errorBody is the standard error response shape.
type errorBody struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// Success sends a 200 success response with the standard envelope.
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, successBody{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 success response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, successBody{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a 200 response including pagination metadata.
func SuccessWithMeta(c *gin.Context, message string, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, successBody{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// ErrorResult pairs an HTTP status with optional field errors, used by the
// central error handler.
type ErrorResult struct {
	Status int
	Msg    string
	Fields []FieldError
}

// AbortWithError aborts the request and writes a standard error envelope.
func AbortWithError(c *gin.Context, er ErrorResult) {
	c.AbortWithStatusJSON(er.Status, errorBody{
		Success: false,
		Message: er.Msg,
		Errors:  er.Fields,
	})
}

// APIError is an error that carries an HTTP status and optional field errors.
// Service-layer errors can be wrapped as APIError to be handled centrally.
type APIError struct {
	Status  int
	Message string
	Fields  []FieldError
}

func (e *APIError) Error() string { return e.Message }

// ErrorHandler is the central error-recovery middleware. Any panic or
// unhandled error recovered during a request is turned into a clean 500
// envelope instead of a raw stack trace.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// If the response has not been written (no handler produced a body),
		// check for unhandled errors in the context.
		if !c.Writer.Written() {
			if len(c.Errors) > 0 {
				err := c.Errors.Last().Err
				var apiErr *APIError
				if errors.As(err, &apiErr) {
					AbortWithError(c, ErrorResult{
						Status: apiErr.Status,
						Msg:    apiErr.Message,
						Fields: apiErr.Fields,
					})
					return
				}
				AbortWithError(c, ErrorResult{
					Status: http.StatusInternalServerError,
					Msg:    "internal server error",
				})
				return
		}
		}
	}
}

// Recovery recovers from panics and logs them, then returns a 500 envelope.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		log.Printf("[PANIC] %v", recovered)
		AbortWithError(c, ErrorResult{
			Status: http.StatusInternalServerError,
			Msg:    "internal server error",
		})
	})
}

// Package response standardizes the JSON envelope returned by all handlers.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Meta carries pagination and other list metadata.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"page_size,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

type successBody struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

// ErrorPayload is the body of every non-2xx response.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type errorBody struct {
	Error ErrorPayload `json:"error"`
}

// OK writes 200 with data.
func OK(c *gin.Context, data any) { c.JSON(http.StatusOK, successBody{Data: data}) }

// Created writes 201 with data.
func Created(c *gin.Context, data any) { c.JSON(http.StatusCreated, successBody{Data: data}) }

// NoContent writes 204.
func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

// List writes 200 with data plus pagination meta.
func List(c *gin.Context, data any, meta *Meta) {
	c.JSON(http.StatusOK, successBody{Data: data, Meta: meta})
}

// Error writes an error envelope and aborts the chain.
func Error(c *gin.Context, status int, code, message string, details ...any) {
	var d any
	if len(details) > 0 {
		d = details[0]
	}
	c.AbortWithStatusJSON(status, errorBody{Error: ErrorPayload{Code: code, Message: message, Details: d}})
}

// Common shortcuts.
func BadRequest(c *gin.Context, msg string, details ...any) {
	Error(c, http.StatusBadRequest, "bad_request", msg, details...)
}
func Unauthorized(c *gin.Context, msg string) { Error(c, http.StatusUnauthorized, "unauthorized", msg) }
func Forbidden(c *gin.Context, msg string)    { Error(c, http.StatusForbidden, "forbidden", msg) }
func NotFound(c *gin.Context, msg string)     { Error(c, http.StatusNotFound, "not_found", msg) }
func Conflict(c *gin.Context, msg string)     { Error(c, http.StatusConflict, "conflict", msg) }
func Locked(c *gin.Context, msg string)       { Error(c, http.StatusLocked, "locked", msg) }
func Internal(c *gin.Context, msg string)     { Error(c, http.StatusInternalServerError, "internal_error", msg) }

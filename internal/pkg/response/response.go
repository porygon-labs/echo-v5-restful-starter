// Package response provides a consistent JSON envelope for all API endpoints.
//
//	{
//	  "meta": { "is_success": true, "message": "ok" },
//	  "data": { ... }
//	}
package response

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Meta holds request-level metadata.
type Meta struct {
	IsSuccess bool   `json:"is_success"`
	Message   string `json:"message"`
	Hint      string `json:"hint,omitempty"`
}

// Envelope is the standard API response wrapper.
type Envelope struct {
	Meta Meta `json:"meta"`
	Data any  `json:"data"`
}

// Success sends a response with data. IsSuccess is derived from the status code
// (true for 1xx–3xx, false for 4xx–5xx).
func Success(c *echo.Context, status int, data any) error {
	return c.JSON(status, Envelope{
		Meta: Meta{IsSuccess: status < 400, Message: http.StatusText(status)},
		Data: data,
	})
}

// Error sends a 4xx/5xx response with an optional hint.
func Error(c *echo.Context, status int, message string, hint ...string) error {
	m := Meta{IsSuccess: false, Message: message}
	if len(hint) > 0 {
		m.Hint = hint[0]
	}
	return c.JSON(status, Envelope{Meta: m, Data: nil})
}

// OK is a shorthand for Success with 200.
func OK(c *echo.Context, data any) error {
	return Success(c, http.StatusOK, data)
}

// Created is a shorthand for Success with 201.
func Created(c *echo.Context, data any) error {
	return Success(c, http.StatusCreated, data)
}

// NoContent sends 204 with no body.
func NoContent(c *echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

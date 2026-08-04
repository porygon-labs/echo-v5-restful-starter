package response_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-restful-api/internal/pkg/response"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/echotest"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func newContext(t *testing.T) (*echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	c, rec := echotest.ContextConfig{
		Request:  httptest.NewRequest(http.MethodGet, "/", nil),
		Response: httptest.NewRecorder(),
	}.ToContextRecorder(t)
	return c, rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return env
}

func TestOK(t *testing.T) {
	c, rec := newContext(t)

	err := response.OK(c, map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("OK returned error: %v", err)
	}

	env := decode(t, rec)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !env.Meta.IsSuccess {
		t.Error("IsSuccess = false, want true")
	}
	if env.Meta.Message != "OK" {
		t.Errorf("Message = %q, want %q", env.Meta.Message, "OK")
	}
	data, _ := env.Data.(map[string]any)
	if data["key"] != "value" {
		t.Errorf("data[key] = %v, want %q", data["key"], "value")
	}
}

func TestCreated(t *testing.T) {
	c, rec := newContext(t)

	err := response.Created(c, "resource")
	if err != nil {
		t.Fatalf("Created returned error: %v", err)
	}

	env := decode(t, rec)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if !env.Meta.IsSuccess {
		t.Error("IsSuccess = false, want true")
	}
	if env.Meta.Message != "Created" {
		t.Errorf("Message = %q, want %q", env.Meta.Message, "Created")
	}
}

func TestSuccess(t *testing.T) {
	c, rec := newContext(t)

	err := response.Success(c, http.StatusAccepted, 42)
	if err != nil {
		t.Fatalf("Success returned error: %v", err)
	}

	env := decode(t, rec)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if !env.Meta.IsSuccess {
		t.Error("IsSuccess = false, want true")
	}
	if env.Meta.Message != "Accepted" {
		t.Errorf("Message = %q, want %q", env.Meta.Message, "Accepted")
	}
	// json numbers decode as float64
	if v, ok := env.Data.(float64); !ok || v != 42 {
		t.Errorf("data = %v, want 42", env.Data)
	}
}

func TestSuccess_WithErrorStatus(t *testing.T) {
	c, rec := newContext(t)

	_ = response.Success(c, http.StatusServiceUnavailable, map[string]string{"db": "down"})

	env := decode(t, rec)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if env.Meta.IsSuccess {
		t.Error("IsSuccess = true, want false for 5xx")
	}
}

func TestError(t *testing.T) {
	c, rec := newContext(t)

	err := response.Error(c, http.StatusNotFound, "book not found")
	if err != nil {
		t.Fatalf("Error returned error: %v", err)
	}

	env := decode(t, rec)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if env.Meta.IsSuccess {
		t.Error("IsSuccess = true, want false")
	}
	if env.Meta.Message != "book not found" {
		t.Errorf("Message = %q, want %q", env.Meta.Message, "book not found")
	}
	if env.Meta.Hint != "" {
		t.Errorf("Hint = %q, want empty", env.Meta.Hint)
	}
	if env.Data != nil {
		t.Errorf("Data = %v, want nil", env.Data)
	}
}

func TestErrorWithHint(t *testing.T) {
	c, rec := newContext(t)

	err := response.Error(c, http.StatusUnprocessableEntity, "validation failed", "name is required")
	if err != nil {
		t.Fatalf("Error returned error: %v", err)
	}

	env := decode(t, rec)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	if env.Meta.Hint != "name is required" {
		t.Errorf("Hint = %q, want %q", env.Meta.Hint, "name is required")
	}
}

func TestNoContent(t *testing.T) {
	c, rec := newContext(t)

	err := response.NoContent(c)
	if err != nil {
		t.Fatalf("NoContent returned error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %s, want empty", rec.Body.String())
	}
}

func TestHintOmittedWhenEmpty(t *testing.T) {
	c, rec := newContext(t)

	_ = response.Error(c, http.StatusBadRequest, "bad")

	var raw map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode raw: %v", err)
	}

	meta, _ := raw["meta"].(map[string]any)
	if _, exists := meta["hint"]; exists {
		t.Error("hint key should be omitted when empty")
	}
}

func TestHintPresent(t *testing.T) {
	c, rec := newContext(t)

	_ = response.Error(c, http.StatusBadRequest, "bad", "do this instead")

	var raw map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode raw: %v", err)
	}

	meta, _ := raw["meta"].(map[string]any)
	if _, exists := meta["hint"]; !exists {
		t.Error("hint key should be present when provided")
	}
}

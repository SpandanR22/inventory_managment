// Package httpapi holds the small helpers every HTTP handler shares:
// reading and writing JSON, error responses, and URL parameters.
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const maxBodyBytes = 1 << 20 // 1 MB

// WriteJSON writes data as a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The status is already sent, so we can only log an encoding failure.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("encoding json response", "error", err)
	}
}

// WriteError writes {"error": message}.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// WriteValidationErrors writes a 422 with one message per invalid field.
func WriteValidationErrors(w http.ResponseWriter, fields map[string]string) {
	WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{
		"error":  "validation failed",
		"fields": fields,
	})
}

// WriteFieldError reports a problem with one field, using the same shape as
// a validation error so the frontend can show it next to that field. Used
// for conflicts such as a duplicate SKU (409).
func WriteFieldError(w http.ResponseWriter, status int, field, message string) {
	WriteJSON(w, status, map[string]any{
		"error":  message,
		"fields": map[string]string{field: message},
	})
}

// ServerError logs the real error and sends the client a generic 500, so
// internal details such as SQL errors never leave the server.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error", "method", r.Method, "path", r.URL.Path, "error", err)
	WriteError(w, http.StatusInternalServerError, "internal server error")
}

// ReadJSON decodes a single JSON object from the request body into dst,
// which must be a pointer. Unknown fields are rejected.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body must not be empty")
		}
		return err
	}

	// A second Decode must hit the end of the body; anything else means
	// the client sent extra data after the object.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

// ParseID reads the {id} URL parameter as a positive integer.
func ParseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id must be a positive integer")
	}
	return id, nil
}

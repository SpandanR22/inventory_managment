package stock

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/SpandanR22/inventory_managment/internal/httpapi"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// List handles GET /stock-movements?product_id=1&limit=20.
// Both query parameters are optional.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var productID *int64
	if raw := query.Get("product_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			httpapi.WriteError(w, http.StatusBadRequest, "product_id must be a positive integer")
			return
		}
		productID = &id
	}

	limit := defaultLimit
	if raw := query.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 || n > maxLimit {
			httpapi.WriteError(w, http.StatusBadRequest, "limit must be between 1 and 200")
			return
		}
		limit = n
	}

	movements, err := h.repo.List(r.Context(), productID, limit)
	if err != nil {
		httpapi.ServerError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"movements": movements})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateMovementRequest
	if err := httpapi.ReadJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	if errs := req.Validate(); len(errs) > 0 {
		httpapi.WriteValidationErrors(w, errs)
		return
	}

	movement, err := h.repo.Create(r.Context(), req)
	switch {
	case errors.Is(err, ErrProductNotFound):
		httpapi.WriteFieldError(w, http.StatusUnprocessableEntity, "product_id", ErrProductNotFound.Error())
	case errors.Is(err, ErrInsufficientStock):
		httpapi.WriteFieldError(w, http.StatusConflict, "quantity", ErrInsufficientStock.Error())
	case err != nil:
		httpapi.ServerError(w, r, err)
	default:
		httpapi.WriteJSON(w, http.StatusCreated, movement)
	}
}

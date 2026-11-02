package products

import (
	"errors"
	"net/http"

	"github.com/SpandanR22/inventory_managment/internal/httpapi"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.ListProducts(r.Context())
	if err != nil {
		httpapi.ServerError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"products": products})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httpapi.ParseID(r)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := httpapi.ReadJSON(w, r, &req); err != nil {
		// Decode errors describe the client's mistake, so they are safe to return.
		httpapi.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		httpapi.WriteValidationErrors(w, errs)
		return
	}

	product, err := h.service.CreateProduct(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, product)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpapi.ParseID(r)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var in ProductInput
	if err := httpapi.ReadJSON(w, r, &in); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if errs := in.Validate(); len(errs) > 0 {
		httpapi.WriteValidationErrors(w, errs)
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), id, in)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpapi.ParseID(r)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeServiceError maps service errors to HTTP status codes. errors.Is
// still finds the sentinel error even though the repository wrapped it.
func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, ErrNotFound.Error())
	case errors.Is(err, ErrDuplicateSKU):
		httpapi.WriteFieldError(w, http.StatusConflict, "sku", ErrDuplicateSKU.Error())
	case errors.Is(err, ErrInvalidReference):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, ErrInvalidReference.Error())
	default:
		httpapi.ServerError(w, r, err)
	}
}

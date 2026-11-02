package suppliers

import (
	"errors"
	"net/http"

	"github.com/SpandanR22/inventory_managment/internal/httpapi"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.repo.List(r.Context())
	if err != nil {
		httpapi.ServerError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"suppliers": suppliers})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	in, ok := readInput(w, r)
	if !ok {
		return
	}

	supplier, err := h.repo.Create(r.Context(), in)
	if err != nil {
		writeRepoError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, supplier)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpapi.ParseID(r)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	in, ok := readInput(w, r)
	if !ok {
		return
	}

	supplier, err := h.repo.Update(r.Context(), id, in)
	if err != nil {
		writeRepoError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, supplier)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpapi.ParseID(r)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeRepoError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// readInput decodes, normalizes and validates the body. If anything is
// wrong it writes the error response itself and returns ok == false.
func readInput(w http.ResponseWriter, r *http.Request) (SupplierInput, bool) {
	var in SupplierInput
	if err := httpapi.ReadJSON(w, r, &in); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return in, false
	}

	in.Normalize()
	if errs := in.Validate(); len(errs) > 0 {
		httpapi.WriteValidationErrors(w, errs)
		return in, false
	}
	return in, true
}

func writeRepoError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, ErrNotFound.Error())
	case errors.Is(err, ErrDuplicateName):
		httpapi.WriteFieldError(w, http.StatusConflict, "name", ErrDuplicateName.Error())
	default:
		httpapi.ServerError(w, r, err)
	}
}

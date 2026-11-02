// Package stock records stock movements. Every movement changes a product's
// quantity, and both writes happen in one database transaction.
package stock

import (
	"errors"
	"time"
	"unicode/utf8"
)

// Movement types.
const (
	TypeIn     = "IN"     // goods received, adds stock
	TypeOut    = "OUT"    // goods sold or shipped, removes stock
	TypeAdjust = "ADJUST" // correction after a stock count, signed
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("not enough stock for this movement")
)

type Movement struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	ProductName  string    `json:"product_name"`
	Type         string    `json:"type"`
	Change       int       `json:"change"`
	BalanceAfter int       `json:"balance_after"`
	Note         *string   `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateMovementRequest is the request body for recording a movement.
// For IN and OUT, Quantity is a positive amount; for ADJUST it is signed
// (for example -2 after finding two damaged items).
type CreateMovementRequest struct {
	ProductID int64   `json:"product_id"`
	Type      string  `json:"type"`
	Quantity  int     `json:"quantity"`
	Note      *string `json:"note"`
}

func (req CreateMovementRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if req.ProductID <= 0 {
		errs["product_id"] = "is required"
	}

	switch req.Type {
	case TypeIn, TypeOut:
		if req.Quantity <= 0 {
			errs["quantity"] = "must be greater than zero"
		}
	case TypeAdjust:
		if req.Quantity == 0 {
			errs["quantity"] = "must not be zero"
		}
	default:
		errs["type"] = "must be IN, OUT or ADJUST"
	}

	if req.Note != nil && utf8.RuneCountInString(*req.Note) > 500 {
		errs["note"] = "must be at most 500 characters"
	}

	return errs
}

// Change returns the signed amount this movement adds to stock.
func (req CreateMovementRequest) Change() int {
	if req.Type == TypeOut {
		return -req.Quantity
	}
	return req.Quantity
}

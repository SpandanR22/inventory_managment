package products

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// Errors the service returns; the handler maps each one to a status code.
var (
	ErrNotFound         = errors.New("product not found")
	ErrDuplicateSKU     = errors.New("a product with this SKU already exists")
	ErrInvalidReference = errors.New("the selected category or supplier does not exist")
)

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	SKU         string    `json:"sku"`
	Description *string   `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	CategoryID  *int64    `json:"category_id"`
	SupplierID  *int64    `json:"supplier_id"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductInput holds the fields a client may set on create and update.
// ID and timestamps come from the database, and quantity only changes
// through stock movements, so none of them are here.
type ProductInput struct {
	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Description *string `json:"description"`
	PriceCents  int64   `json:"price_cents"`
	CategoryID  *int64  `json:"category_id"`
	SupplierID  *int64  `json:"supplier_id"`
}

// CreateProductRequest embeds ProductInput, so its fields appear at the top
// level of the JSON, and adds the opening stock quantity.
type CreateProductRequest struct {
	ProductInput
	Quantity int `json:"quantity"`
}

// Normalize trims whitespace and turns an empty description into nil.
// It has a pointer receiver because it modifies the input.
func (in *ProductInput) Normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.SKU = strings.TrimSpace(in.SKU)
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		if d == "" {
			in.Description = nil
		} else {
			in.Description = &d
		}
	}
}

// Validate returns one message per invalid field; an empty map means valid.
// It has a value receiver because it only reads.
func (in ProductInput) Validate() map[string]string {
	errs := make(map[string]string)

	if strings.TrimSpace(in.Name) == "" {
		errs["name"] = "is required"
	} else if utf8.RuneCountInString(in.Name) > 200 {
		errs["name"] = "must be at most 200 characters"
	}

	if strings.TrimSpace(in.SKU) == "" {
		errs["sku"] = "is required"
	} else if utf8.RuneCountInString(in.SKU) > 100 {
		errs["sku"] = "must be at most 100 characters"
	}

	if in.PriceCents <= 0 {
		errs["price_cents"] = "must be a positive integer"
	}

	if in.Description != nil && utf8.RuneCountInString(*in.Description) > 500 {
		errs["description"] = "must be at most 500 characters"
	}

	if in.CategoryID != nil && *in.CategoryID <= 0 {
		errs["category_id"] = "must be a positive integer"
	}

	if in.SupplierID != nil && *in.SupplierID <= 0 {
		errs["supplier_id"] = "must be a positive integer"
	}

	return errs
}

func (req CreateProductRequest) Validate() map[string]string {
	errs := req.ProductInput.Validate()
	if req.Quantity < 0 {
		errs["quantity"] = "must be a non-negative integer"
	}
	return errs
}

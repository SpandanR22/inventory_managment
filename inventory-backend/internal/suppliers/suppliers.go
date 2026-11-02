// Package suppliers manages the companies products are bought from. Like
// categories, it uses only a handler and a repository.
package suppliers

import (
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrNotFound      = errors.New("supplier not found")
	ErrDuplicateName = errors.New("a supplier with this name already exists")
)

type Supplier struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     *string   `json:"email"`
	Phone     *string   `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SupplierInput is the request body for create and update.
type SupplierInput struct {
	Name  string  `json:"name"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
}

func (in *SupplierInput) Normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = trimOptional(in.Email)
	in.Phone = trimOptional(in.Phone)
}

func (in SupplierInput) Validate() map[string]string {
	errs := make(map[string]string)

	if in.Name == "" {
		errs["name"] = "is required"
	} else if utf8.RuneCountInString(in.Name) > 200 {
		errs["name"] = "must be at most 200 characters"
	}

	if in.Email != nil {
		if _, err := mail.ParseAddress(*in.Email); err != nil {
			errs["email"] = "must be a valid email address"
		}
	}

	if in.Phone != nil && utf8.RuneCountInString(*in.Phone) > 30 {
		errs["phone"] = "must be at most 30 characters"
	}

	return errs
}

// trimOptional trims an optional string and turns "" into nil, so the
// database stores NULL rather than an empty string.
func trimOptional(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

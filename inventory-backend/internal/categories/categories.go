// Package categories manages product categories. It has no business rules
// beyond validation, so it uses only a handler and a repository: a service
// layer here would just forward every call.
package categories

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrNotFound      = errors.New("category not found")
	ErrDuplicateName = errors.New("a category with this name already exists")
)

type Category struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CategoryInput is the request body for create and update.
type CategoryInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (in *CategoryInput) Normalize() {
	in.Name = strings.TrimSpace(in.Name)
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		if d == "" {
			in.Description = nil
		} else {
			in.Description = &d
		}
	}
}

func (in CategoryInput) Validate() map[string]string {
	errs := make(map[string]string)

	if in.Name == "" {
		errs["name"] = "is required"
	} else if utf8.RuneCountInString(in.Name) > 100 {
		errs["name"] = "must be at most 100 characters"
	}

	if in.Description != nil && utf8.RuneCountInString(*in.Description) > 500 {
		errs["description"] = "must be at most 500 characters"
	}

	return errs
}

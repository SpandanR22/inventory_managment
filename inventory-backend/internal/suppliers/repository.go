package suppliers

import (
	"context"
	"errors"
	"fmt"

	"github.com/SpandanR22/inventory_managment/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const supplierColumns = `id, name, email, phone, created_at, updated_at`

func scanSupplier(row pgx.Row) (Supplier, error) {
	var s Supplier
	err := row.Scan(&s.ID, &s.Name, &s.Email, &s.Phone, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) List(ctx context.Context) ([]Supplier, error) {
	rows, err := r.db.Query(ctx, `SELECT `+supplierColumns+` FROM suppliers ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("querying suppliers: %w", err)
	}
	defer rows.Close()

	suppliers := []Supplier{}
	for rows.Next() {
		s, err := scanSupplier(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning supplier: %w", err)
		}
		suppliers = append(suppliers, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating suppliers: %w", err)
	}
	return suppliers, nil
}

func (r *Repository) Create(ctx context.Context, in SupplierInput) (Supplier, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO suppliers (name, email, phone) VALUES ($1, $2, $3) RETURNING `+supplierColumns,
		in.Name, in.Email, in.Phone)
	s, err := scanSupplier(row)
	if err != nil {
		return Supplier{}, fmt.Errorf("inserting supplier: %w", translateError(err))
	}
	return s, nil
}

func (r *Repository) Update(ctx context.Context, id int64, in SupplierInput) (Supplier, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE suppliers SET name = $1, email = $2, phone = $3, updated_at = now()
		 WHERE id = $4 RETURNING `+supplierColumns,
		in.Name, in.Email, in.Phone, id)
	s, err := scanSupplier(row)
	if err != nil {
		return Supplier{}, fmt.Errorf("updating supplier %d: %w", id, translateError(err))
	}
	return s, nil
}

// Delete removes a supplier. Its products keep existing with no supplier
// (ON DELETE SET NULL).
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM suppliers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting supplier %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func translateError(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNotFound
	case database.IsUniqueViolation(err):
		return ErrDuplicateName
	default:
		return err
	}
}

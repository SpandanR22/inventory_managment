package categories

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

const categoryColumns = `id, name, description, created_at, updated_at`

func scanCategory(row pgx.Row) (Category, error) {
	var c Category
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *Repository) List(ctx context.Context) ([]Category, error) {
	rows, err := r.db.Query(ctx, `SELECT `+categoryColumns+` FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("querying categories: %w", err)
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating categories: %w", err)
	}
	return categories, nil
}

func (r *Repository) Create(ctx context.Context, in CategoryInput) (Category, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING `+categoryColumns,
		in.Name, in.Description)
	c, err := scanCategory(row)
	if err != nil {
		return Category{}, fmt.Errorf("inserting category: %w", translateError(err))
	}
	return c, nil
}

func (r *Repository) Update(ctx context.Context, id int64, in CategoryInput) (Category, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE categories SET name = $1, description = $2, updated_at = now()
		 WHERE id = $3 RETURNING `+categoryColumns,
		in.Name, in.Description, id)
	c, err := scanCategory(row)
	if err != nil {
		return Category{}, fmt.Errorf("updating category %d: %w", id, translateError(err))
	}
	return c, nil
}

// Delete removes a category. Products in it keep existing with no category,
// because the foreign key is declared ON DELETE SET NULL.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting category %d: %w", id, err)
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

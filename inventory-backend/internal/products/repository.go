package products

import (
	"context"
	"errors"
	"fmt"

	"github.com/SpandanR22/inventory_managment/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListProducts(ctx context.Context) ([]Product, error)
	GetProduct(ctx context.Context, id int64) (Product, error)
	CreateProduct(ctx context.Context, req CreateProductRequest) (Product, error)
	UpdateProduct(ctx context.Context, id int64, in ProductInput) (Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

// productColumns is shared by every query so Scan always sees the same order.
const productColumns = `id, name, sku, description, price_cents, category_id, supplier_id, quantity, created_at, updated_at`

// scanProduct works for both a single row (QueryRow) and a row inside a
// loop (Query), because both have a Scan method and so satisfy pgx.Row.
func scanProduct(row pgx.Row) (Product, error) {
	var p Product
	err := row.Scan(
		&p.ID,
		&p.Name,
		&p.SKU,
		&p.Description,
		&p.PriceCents,
		&p.CategoryID,
		&p.SupplierID,
		&p.Quantity,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	return p, err
}

func (r *postgresRepository) ListProducts(ctx context.Context) ([]Product, error) {
	rows, err := r.db.Query(ctx, `SELECT `+productColumns+` FROM products ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("querying products: %w", err)
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating products: %w", err)
	}

	return products, nil
}

func (r *postgresRepository) GetProduct(ctx context.Context, id int64) (Product, error) {
	row := r.db.QueryRow(ctx, `SELECT `+productColumns+` FROM products WHERE id = $1`, id)
	p, err := scanProduct(row)
	if err != nil {
		return Product{}, fmt.Errorf("getting product %d: %w", id, translateError(err))
	}
	return p, nil
}

func (r *postgresRepository) CreateProduct(ctx context.Context, req CreateProductRequest) (Product, error) {
	const query = `
		INSERT INTO products (name, sku, description, price_cents, category_id, supplier_id, quantity)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + productColumns

	row := r.db.QueryRow(ctx, query,
		req.Name,
		req.SKU,
		req.Description,
		req.PriceCents,
		req.CategoryID,
		req.SupplierID,
		req.Quantity,
	)
	p, err := scanProduct(row)
	if err != nil {
		return Product{}, fmt.Errorf("inserting product: %w", translateError(err))
	}
	return p, nil
}

func (r *postgresRepository) UpdateProduct(ctx context.Context, id int64, in ProductInput) (Product, error) {
	const query = `
		UPDATE products
		SET name = $1, sku = $2, description = $3, price_cents = $4,
		    category_id = $5, supplier_id = $6, updated_at = now()
		WHERE id = $7
		RETURNING ` + productColumns

	row := r.db.QueryRow(ctx, query,
		in.Name,
		in.SKU,
		in.Description,
		in.PriceCents,
		in.CategoryID,
		in.SupplierID,
		id,
	)
	p, err := scanProduct(row)
	if err != nil {
		return Product{}, fmt.Errorf("updating product %d: %w", id, translateError(err))
	}
	return p, nil
}

func (r *postgresRepository) DeleteProduct(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting product %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// translateError turns database errors into this package's errors, so the
// layers above never need to know about pgx or PostgreSQL error codes.
func translateError(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNotFound
	case database.IsUniqueViolation(err):
		return ErrDuplicateSKU
	case database.IsForeignKeyViolation(err):
		return ErrInvalidReference
	default:
		return err
	}
}

package stock

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// List returns the newest movements first. If productID is nil, movements
// for all products are returned.
func (r *Repository) List(ctx context.Context, productID *int64, limit int) ([]Movement, error) {
	const query = `
		SELECT m.id, m.product_id, p.name, m.type, m.change, m.balance_after, m.note, m.created_at
		FROM stock_movements m
		JOIN products p ON p.id = m.product_id
		WHERE ($1::bigint IS NULL OR m.product_id = $1)
		ORDER BY m.id DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, productID, limit)
	if err != nil {
		return nil, fmt.Errorf("querying stock movements: %w", err)
	}
	defer rows.Close()

	movements := []Movement{}
	for rows.Next() {
		var m Movement
		if err := rows.Scan(
			&m.ID,
			&m.ProductID,
			&m.ProductName,
			&m.Type,
			&m.Change,
			&m.BalanceAfter,
			&m.Note,
			&m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning stock movement: %w", err)
		}
		movements = append(movements, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating stock movements: %w", err)
	}
	return movements, nil
}

// Create records a movement and updates the product quantity atomically:
// either both writes happen or neither does.
func (r *Repository) Create(ctx context.Context, req CreateMovementRequest) (Movement, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Movement{}, fmt.Errorf("starting transaction: %w", err)
	}
	// Rollback after a successful Commit does nothing, so this is safe and
	// guarantees the transaction is closed on every error path.
	defer tx.Rollback(ctx)

	// FOR UPDATE locks the product row until the transaction ends. Two
	// requests selling the last item at the same moment are processed one
	// after the other, so the second sees the updated quantity.
	var name string
	var current int
	err = tx.QueryRow(ctx,
		`SELECT name, quantity FROM products WHERE id = $1 FOR UPDATE`,
		req.ProductID,
	).Scan(&name, &current)
	if errors.Is(err, pgx.ErrNoRows) {
		return Movement{}, ErrProductNotFound
	}
	if err != nil {
		return Movement{}, fmt.Errorf("locking product %d: %w", req.ProductID, err)
	}

	balance := current + req.Change()
	if balance < 0 {
		return Movement{}, ErrInsufficientStock
	}

	if _, err := tx.Exec(ctx,
		`UPDATE products SET quantity = $1, updated_at = now() WHERE id = $2`,
		balance, req.ProductID,
	); err != nil {
		return Movement{}, fmt.Errorf("updating quantity: %w", err)
	}

	m := Movement{ProductName: name}
	err = tx.QueryRow(ctx,
		`INSERT INTO stock_movements (product_id, type, change, balance_after, note)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, product_id, type, change, balance_after, note, created_at`,
		req.ProductID, req.Type, req.Change(), balance, req.Note,
	).Scan(&m.ID, &m.ProductID, &m.Type, &m.Change, &m.BalanceAfter, &m.Note, &m.CreatedAt)
	if err != nil {
		return Movement{}, fmt.Errorf("inserting stock movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Movement{}, fmt.Errorf("committing transaction: %w", err)
	}
	return m, nil
}

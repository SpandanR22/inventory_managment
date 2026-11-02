// Package database owns the connection pool, the schema, and helpers for
// recognising PostgreSQL errors.
package database

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

// Connect creates a connection pool and checks that the database is reachable.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	// pgxpool.New is lazy; Ping is the first real connection attempt.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}

// Migrate applies schema.sql. It is safe to run on every startup.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("applying schema: %w", err)
	}
	return nil
}

// PostgreSQL error codes, see https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	codeForeignKeyViolation = "23503"
	codeUniqueViolation     = "23505"
)

// IsUniqueViolation reports whether err was caused by a UNIQUE constraint.
func IsUniqueViolation(err error) bool {
	return hasCode(err, codeUniqueViolation)
}

// IsForeignKeyViolation reports whether err was caused by a FOREIGN KEY
// constraint, for example a product pointing at a category that does not exist.
func IsForeignKeyViolation(err error) bool {
	return hasCode(err, codeForeignKeyViolation)
}

func hasCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

// Package dashboard serves read-only summary figures and reports. It only
// reads, so it queries the tables directly instead of going through the
// other packages.
package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SpandanR22/inventory_managment/internal/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultLowStockThreshold = 10

type Totals struct {
	Products        int64 `json:"products"`
	Categories      int64 `json:"categories"`
	Suppliers       int64 `json:"suppliers"`
	Units           int64 `json:"units"`
	StockValueCents int64 `json:"stock_value_cents"`
}

type LowStockItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// CategoryReport is one row of the stock value report.
type CategoryReport struct {
	Category        string `json:"category"`
	Products        int64  `json:"products"`
	Units           int64  `json:"units"`
	StockValueCents int64  `json:"stock_value_cents"`
}

type Summary struct {
	Totals            Totals           `json:"totals"`
	LowStockThreshold int              `json:"low_stock_threshold"`
	LowStock          []LowStockItem   `json:"low_stock"`
	ByCategory        []CategoryReport `json:"by_category"`
}

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

// Get handles GET /dashboard?low_stock_threshold=10.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	threshold := defaultLowStockThreshold
	if raw := r.URL.Query().Get("low_stock_threshold"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			httpapi.WriteError(w, http.StatusBadRequest, "low_stock_threshold must be a non-negative integer")
			return
		}
		threshold = n
	}

	summary, err := h.summary(r.Context(), threshold)
	if err != nil {
		httpapi.ServerError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, summary)
}

func (h *Handler) summary(ctx context.Context, threshold int) (Summary, error) {
	s := Summary{LowStockThreshold: threshold}

	// sum() of a bigint returns NUMERIC in Postgres, so cast back to bigint
	// to scan into int64. COALESCE turns "no rows" (NULL) into 0.
	err := h.db.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM products),
			(SELECT count(*) FROM categories),
			(SELECT count(*) FROM suppliers),
			(SELECT COALESCE(sum(quantity), 0)::bigint FROM products),
			(SELECT COALESCE(sum(quantity::bigint * price_cents), 0)::bigint FROM products)`,
	).Scan(&s.Totals.Products, &s.Totals.Categories, &s.Totals.Suppliers, &s.Totals.Units, &s.Totals.StockValueCents)
	if err != nil {
		return Summary{}, fmt.Errorf("querying totals: %w", err)
	}

	if s.LowStock, err = h.lowStock(ctx, threshold); err != nil {
		return Summary{}, err
	}
	if s.ByCategory, err = h.byCategory(ctx); err != nil {
		return Summary{}, err
	}
	return s, nil
}

func (h *Handler) lowStock(ctx context.Context, threshold int) ([]LowStockItem, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, name, sku, quantity FROM products WHERE quantity <= $1 ORDER BY quantity, name`,
		threshold)
	if err != nil {
		return nil, fmt.Errorf("querying low stock: %w", err)
	}
	defer rows.Close()

	items := []LowStockItem{}
	for rows.Next() {
		var it LowStockItem
		if err := rows.Scan(&it.ID, &it.Name, &it.SKU, &it.Quantity); err != nil {
			return nil, fmt.Errorf("scanning low stock item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (h *Handler) byCategory(ctx context.Context) ([]CategoryReport, error) {
	rows, err := h.db.Query(ctx, `
		SELECT COALESCE(c.name, 'Uncategorized'),
		       count(p.id),
		       COALESCE(sum(p.quantity), 0)::bigint,
		       COALESCE(sum(p.quantity::bigint * p.price_cents), 0)::bigint
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		GROUP BY c.name
		ORDER BY 4 DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying category report: %w", err)
	}
	defer rows.Close()

	reports := []CategoryReport{}
	for rows.Next() {
		var cr CategoryReport
		if err := rows.Scan(&cr.Category, &cr.Products, &cr.Units, &cr.StockValueCents); err != nil {
			return nil, fmt.Errorf("scanning category report: %w", err)
		}
		reports = append(reports, cr)
	}
	return reports, rows.Err()
}

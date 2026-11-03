package main

import (
	"net/http"
	"time"

	"github.com/SpandanR22/inventory_managment/internal/categories"
	"github.com/SpandanR22/inventory_managment/internal/dashboard"
	"github.com/SpandanR22/inventory_managment/internal/httpapi"
	"github.com/SpandanR22/inventory_managment/internal/products"
	"github.com/SpandanR22/inventory_managment/internal/stock"
	"github.com/SpandanR22/inventory_managment/internal/suppliers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// routes builds every dependency and registers every endpoint, so this one
// function is the map of the whole API.
func routes(db *pgxpool.Pool) http.Handler {
	productHandler := products.NewHandler(products.NewService(products.NewRepository(db)))
	categoryHandler := categories.NewHandler(categories.NewRepository(db))
	supplierHandler := suppliers.NewHandler(suppliers.NewRepository(db))
	stockHandler := stock.NewHandler(stock.NewRepository(db))
	dashboardHandler := dashboard.NewHandler(db)

	r := chi.NewRouter()
	r.Use(middleware.RequestID) // gives each request an ID for the logs
	r.Use(middleware.Logger)    // logs "GET /api/v1/products 200 3ms"
	r.Use(middleware.Recoverer) // turns a panic into a 500 instead of crashing
	r.Use(middleware.Timeout(10 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/dashboard", dashboardHandler.Get)

		r.Get("/products", productHandler.List)
		r.Post("/products", productHandler.Create)
		r.Get("/products/{id}", productHandler.Get)
		r.Put("/products/{id}", productHandler.Update)
		r.Delete("/products/{id}", productHandler.Delete)

		r.Get("/categories", categoryHandler.List)
		r.Post("/categories", categoryHandler.Create)
		r.Put("/categories/{id}", categoryHandler.Update)
		r.Delete("/categories/{id}", categoryHandler.Delete)

		r.Get("/suppliers", supplierHandler.List)
		r.Post("/suppliers", supplierHandler.Create)
		r.Put("/suppliers/{id}", supplierHandler.Update)
		r.Delete("/suppliers/{id}", supplierHandler.Delete)

		r.Get("/stock-movements", stockHandler.List)
		r.Post("/stock-movements", stockHandler.Create)
	})

	return r
}

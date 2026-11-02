package products

import "context"

type Service interface {
	ListProducts(ctx context.Context) ([]Product, error)
	GetProduct(ctx context.Context, id int64) (Product, error)
	CreateProduct(ctx context.Context, req CreateProductRequest) (Product, error)
	UpdateProduct(ctx context.Context, id int64, in ProductInput) (Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type svc struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &svc{repo: repo}
}

func (s *svc) ListProducts(ctx context.Context) ([]Product, error) {
	return s.repo.ListProducts(ctx)
}

func (s *svc) GetProduct(ctx context.Context, id int64) (Product, error) {
	return s.repo.GetProduct(ctx, id)
}

// CreateProduct stores clean data: req is a copy, so normalizing it here
// does not change the handler's value.
func (s *svc) CreateProduct(ctx context.Context, req CreateProductRequest) (Product, error) {
	req.Normalize()
	return s.repo.CreateProduct(ctx, req)
}

func (s *svc) UpdateProduct(ctx context.Context, id int64, in ProductInput) (Product, error) {
	in.Normalize()
	return s.repo.UpdateProduct(ctx, id, in)
}

func (s *svc) DeleteProduct(ctx context.Context, id int64) error {
	return s.repo.DeleteProduct(ctx, id)
}

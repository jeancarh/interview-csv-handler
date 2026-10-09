package catalog

import (
	"context"
	"time"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

func (s *Service) Create(ctx context.Context, d Draft) (Product, error) {
	p, err := NewProduct(d, s.now())
	if err != nil {
		return Product{}, err
	}
	if err := s.repo.Create(ctx, &p); err != nil {
		return Product{}, err
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id int64, d Draft) (Product, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return Product{}, err
	}
	if err := p.Update(d, s.now()); err != nil {
		return Product{}, err
	}
	if err := s.repo.Update(ctx, &p); err != nil {
		return Product{}, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, id int64) (Product, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id, s.now())
}

func (s *Service) Search(ctx context.Context, q SearchQuery) (Page, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}
	return s.repo.Search(ctx, q)
}

func (s *Service) Categories(ctx context.Context) ([]string, error) {
	return s.repo.Categories(ctx)
}

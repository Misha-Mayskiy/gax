package service

import (
	"context"
	"main/internal/domain"
	esrepo "main/internal/repository/es"

	"github.com/rs/zerolog"
)

type SearchService struct {
	index esrepoInterface
}

type esrepoInterface interface {
	Search(ctx context.Context, q domain.SearchQuery) ([]domain.SearchResult, error)
}

func NewSearchService(indexRepo *esrepo.Repo, log zerolog.Logger) *SearchService {
	return &SearchService{index: indexRepo}
}

func (s *SearchService) Search(ctx context.Context, q domain.SearchQuery) ([]domain.SearchResult, error) {
	// тут можно добавить бизнес-валидацию, нормализацию запроса и аналитику
	if q.Limit <= 0 {
		q.Limit = 20
	}
	return s.index.Search(ctx, q)
}

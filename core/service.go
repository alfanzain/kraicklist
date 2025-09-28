package core

import (
	"context"
	"fmt"
)

type Service interface {
	GetSearch(
		ctx context.Context,
		q string,
		perPage,
		page int,
	) (interface{}, error)
}

type ServiceConfig struct {
	SearchEngine SearchEngine `validate:"nonnil"`
}

type service struct {
	ServiceConfig
}

func NewService(cfg ServiceConfig) (*service, error) {
	return &service{
		ServiceConfig: cfg,
	}, nil
}

func (s *service) GetSearch(
	ctx context.Context,
	q string,
	perPage,
	page int,
) (interface{}, error) {
	searchResult, err := s.SearchEngine.Search(ctx, q, perPage, page)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	fmt.Printf("Found %d results\n", *searchResult.Found)

	return searchResult, nil
}

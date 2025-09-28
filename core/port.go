package core

import (
	"context"

	"github.com/typesense/typesense-go/v3/typesense/api"
)

type SearchEngine interface {
	Search(ctx context.Context, query string, perPage int, page int) (*api.SearchResult, error)
}

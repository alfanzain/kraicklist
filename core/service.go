package core

type Service interface {
	GetSearch(q, queryBy, excludeFields string, perPage, page int) (interface{}, error)
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

func (s *service) GetSearch(q, queryBy, excludeFields string, perPage, page int) (interface{}, error) {
	return s.SearchEngine.Search(q, queryBy, excludeFields, perPage, page)
}

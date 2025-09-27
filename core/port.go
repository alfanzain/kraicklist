package core

type SearchEngine interface {
	Search(q, queryBy, excludeFields string, perPage, page int) (interface{}, error)
}

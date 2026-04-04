package repositories

type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

func NewPagination(page, limit int) Pagination {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

type PageResult[T any] struct {
	Items []T
	Page  int
	Limit int
	Total int
}

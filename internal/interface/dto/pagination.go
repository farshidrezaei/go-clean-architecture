package dto

type PageMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type PaginatedResponse[T any] struct {
	Items []T      `json:"items"`
	Meta  PageMeta `json:"meta"`
}

package types

// CommonFilters contains common filter fields used across entities
type CommonFilters struct {
    SearchTerm *string `form:"searchTerm"`
    Page       int     `form:"page,default=1"`
    Limit      int     `form:"limit,default=10"`
    SortBy     string  `form:"sortBy,default=created_at"`
    SortOrder  string  `form:"sortOrder,default=asc"`
}

// PaginationResponse represents common pagination metadata
type PaginationResponse struct {
    Page      int   `json:"page"`
    Limit     int   `json:"limit"`
    Count     int64 `json:"count"`
    TotalPage int   `json:"totalPage"`
}

// PagedResponse is a generic type for paginated responses
type PagedResponse[T any] struct {
    Meta PaginationResponse `json:"meta"`
    Data []T               `json:"data"`
}

// SearchFilters contains fields for searching entities
type SearchFilters struct {
    CommonFilters
    Ip    *string `form:"ip"`
    Title *string `form:"title"`
}
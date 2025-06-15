package response

// PaginatedResponse represents a paginated response
type PaginatedResponse[T any] struct {
	Items []T        `json:"items"`
	Meta  Pagination `json:"pagination"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// PaginationRequest represents pagination parameters from request
type PaginationRequest struct {
	Page    int `json:"page" form:"page" binding:"min=1"`
	PerPage int `json:"per_page" form:"per_page" binding:"min=1,max=100"`
}

// NewPagination creates a new Pagination instance
func NewPagination(page, perPage, total int) Pagination {
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	return Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse[T any](items []T, page, perPage, total int) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Items: items,
		Meta:  NewPagination(page, perPage, total),
	}
}

// NewPaginatedAPIResponse creates a paginated API response
func NewPaginatedAPIResponse[T any](items []T, page, perPage, total int) *APIResponse[*PaginatedResponse[T]] {
	paginatedData := NewPaginatedResponse(items, page, perPage, total)

	// Convert to Meta for APIResponse
	meta := &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: paginatedData.Meta.TotalPages,
	}

	return SuccessWithMeta(paginatedData, meta)
}

// GetDefaultPagination returns default pagination values
func GetDefaultPagination() PaginationRequest {
	return PaginationRequest{
		Page:    1,
		PerPage: 20,
	}
}

// Validate validates pagination parameters
func (p *PaginationRequest) Validate() error {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	return nil
}

// GetOffset calculates the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// GetLimit returns the limit for database queries
func (p *PaginationRequest) GetLimit() int {
	return p.PerPage
}

package domain

import "database/sql"

// Pagination — метаданные пагинации для ответов со списками.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewPagination вычисляет метаданные пагинации, включая количество страниц.
func NewPagination(page, limit, total int) Pagination {
	totalPages := 0
	if limit > 0 {
		// Округление вверх: (total + limit - 1) / limit.
		totalPages = (total + limit - 1) / limit
	}
	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ErrorResponse — единый формат ошибок API.
type ErrorResponse struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []ErrorDetail `json:"details,omitempty"`
	RequestID string        `json:"request_id,omitempty"`
}

// ErrorDetail — деталь ошибки валидации по конкретному полю.
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// --- Помощники для разворачивания sql.Null* в указатели ---
// nil-указатель соответствует SQL NULL, не-nil — реальному значению.

func nullStringToPtr(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	return &n.String
}

func nullInt32ToPtr(n sql.NullInt32) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int32)
	return &v
}

func nullFloat64ToPtr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	return &n.Float64
}

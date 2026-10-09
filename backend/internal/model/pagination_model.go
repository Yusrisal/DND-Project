package model

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Query carries the pagination, search and sort options shared by every list endpoint.
type Query struct {
	Page     int
	PageSize int
	Search   string
	SortBy   string
	SortDir  string
}

// ParseQuery reads list options from the request, clamping anything out of range.
// allowedSort guards the ORDER BY column, which is interpolated into SQL.
func ParseQuery(r *http.Request, allowedSort []string, defaultSort string) Query {
	q := Query{
		Page:     1,
		PageSize: defaultPageSize,
		Search:   strings.TrimSpace(r.URL.Query().Get("search")),
		SortBy:   defaultSort,
		SortDir:  "desc",
	}

	if v, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && v > 0 {
		q.Page = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && v > 0 {
		q.PageSize = min(v, maxPageSize)
	}
	if v := r.URL.Query().Get("sort_by"); v != "" {
		for _, allowed := range allowedSort {
			if v == allowed {
				q.SortBy = v
				break
			}
		}
	}
	if strings.EqualFold(r.URL.Query().Get("sort_dir"), "asc") {
		q.SortDir = "asc"
	}
	return q
}

func (q Query) Offset() int { return (q.Page - 1) * q.PageSize }

type Page[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func NewPage[T any](items []T, total int64, q Query) Page[T] {
	if items == nil {
		items = []T{}
	}
	totalPages := int((total + int64(q.PageSize) - 1) / int64(q.PageSize))
	return Page[T]{Items: items, Page: q.Page, PageSize: q.PageSize, Total: total, TotalPages: totalPages}
}

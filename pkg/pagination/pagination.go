// Package pagination parses list query params and builds response metadata.
package pagination

import (
	"strconv"

	"github.com/collegeassess/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Params captures common list query parameters.
type Params struct {
	Page     int
	PageSize int
	Search   string
	Sort     string
	Order    string // asc | desc
}

// Offset returns the SQL offset for the current page.
func (p Params) Offset() int { return (p.Page - 1) * p.PageSize }

// OrderClause returns a safe "column direction" string using an allowlist.
func (p Params) OrderClause(allowed map[string]bool, def string) string {
	col := def
	if p.Sort != "" && allowed[p.Sort] {
		col = p.Sort
	}
	dir := "desc"
	if p.Order == "asc" {
		dir = "asc"
	}
	return col + " " + dir
}

// Parse extracts pagination params with sane defaults and bounds.
func Parse(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return Params{
		Page:     page,
		PageSize: size,
		Search:   c.Query("search"),
		Sort:     c.Query("sort"),
		Order:    c.Query("order"),
	}
}

// Meta builds response metadata from total count.
func (p Params) Meta(total int64) *response.Meta {
	pages := int((total + int64(p.PageSize) - 1) / int64(p.PageSize))
	return &response.Meta{Page: p.Page, PageSize: p.PageSize, Total: total, TotalPages: pages}
}

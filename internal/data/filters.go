package data

import (
	"strings"

	"greenlight.camphopkins.com/internal/validator"
	"slices"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be less than or equal to ten million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be less than or equal to one hundred")

	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

func (f Filters) sortColumn() string {
	if slices.Contains(f.SortSafelist, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}

	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
  if strings.HasPrefix(f.Sort, "-") {
    return "DESC"
  }

  return "ASC"
}

func (f Filters) limit() int {
  return f.PageSize
}

func (f Filters) offset() int {
  return (f.Page - 1) * f.PageSize
}


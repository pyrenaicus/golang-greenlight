package data

import (
	"slices"
	"strings"

	"greenlight.cnoua.org/internal/validator"
)

// Add a SortSafelist field to hold the supported sort values
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// Metadata struct for holding pagination metadata
type Metadata struct {
	CurrentPage  int `json:"current_page,omitzero"`
	PageSize     int `json:"page_size,omitzero"`
	FirstPage    int `json:"first_page,omitzero"`
	LastPage     int `json:"last_page,omitzero"`
	TotalRecords int `json:"total_records,omitzero"`
}

// calculateMetadata() returns the appropriate pagination metadata values
// given the total number of records, current page, and page size values.
// Note that when the last page value is calculated we are dividing two int
// values, and the result will also be an int type, with the modulus or
// remainder dropped.
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		// Return an empty Metadata struct if there are no records.
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}

// sortColumn() checks that the client-provided Sort field matches one of the
// entries in our safelist and if it does, extract the column name from the
// Sort field by stripping the leading hyphen character, if it exists.
func (f Filters) sortColumn() string {
	if slices.Contains(f.SortSafelist, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}

	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection() returns the dort direction depending on the prefix character
// of the Sort field.
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

func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page_size parameters contain sensible values.
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Check that the sort parameter matches a value in the safelist.
	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

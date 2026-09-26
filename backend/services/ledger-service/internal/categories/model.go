package categories

import "time"

type Category struct {
	ID        string
	Name      string
	Type      string
	Icon      *string
	Color     *string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

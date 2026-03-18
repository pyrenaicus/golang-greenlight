package data

import (
	"time"
)

type Movie struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int       `json:"year,omitzero"`
	Runtime   int       `json:"runtime,omitzero"` // movie runtime in minutes
	Genres    []string  `json:"genres,omitzero"`
	Version   int       `json:"version"` // starts at 1 and increments each
	// time the movie information is updated
}

package data

import (
	"time"
)

type Movie struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Year      int       `json:"year"`
	Runtime   int       `json:"runtime"` // movie runtime in minutes
	Genres    []string  `json:"genres"`
	Version   int       `json:"version"` // starts at 1 and increments each
	// time the movie information is updated
}

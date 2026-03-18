package data

import (
	"time"
)

type Movie struct {
	ID        int
	CreatedAt time.Time
	Title     string
	Year      int
	Runtime   int // movie runtime in minutes
	Genres    []string
	Version   int // starts at 1 and increments each
	// time the movie information is updated
}

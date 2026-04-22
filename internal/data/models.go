package data

import (
	"database/sql"
	"errors"
)

// ErrRecordNotFound, we'll return this from Get() when looking up a movie
// that doesn't exist on our database.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// create a Models struct which wraps all our models.
type Models struct {
	Movies MovieModel
}

// for ease of use, we add a New() method which returns a Models struct
// containing the initialized MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}

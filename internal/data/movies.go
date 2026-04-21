package data

import (
	"database/sql"
	"errors"
	"time"

	"greenlight.cnoua.org/internal/validator"

	"github.com/lib/pq" // PostgreSQL driver for database/sql
)

// MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// CRUD placeholder methods
// Insert() accepts a pointer to a Movie struct, which should contain the data
// for the new record.
func (m MovieModel) Insert(movie *Movie) error {
	// define a SQL query which inserts a new record in the movies table
	// and returns the system generated data.
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	// create an args array containing the values for the placeholder parameters.
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the elements of the args slice as variadic arguments and scanning
	// the system-generated values into the movie struct.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int) (*Movie, error) {
	// The PostgreSQL id value we use for movie ID starts auto-incrementing at 1
	// by default, so we know that no movies will have ID values less than that.
	// To avoid making an unnecessary database call, we take a shortcut and return
	// an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving the movie data.
	query := `
					SELECT id, created_at, title, year, runtime, genres, version
					FROM movies
					WHERE id = $1
					`

	// Declare a Movie struct to hold the data returned by the query.
	var movie Movie

	// Execute the query using the QueryRow() method, passing in the provided id
	// value as a placeholder parameter, and scan the response data into the
	// fields of the Movie struct. We need to convert the scan target for the
	// genres column using the pq.Array() adapter function.
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	// Handle any errors. If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return ErrRecordNotFound error.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Otherwise, return a pointer to the Movie struct.
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	// Declare the SQL query for updating the record and returning the new version
	// number.
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version`

	// Create an args slice containing the values for the placeholder parameters.
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice
	// as a variadic parameter and scanning the new version value into the movie
	// struct.
	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

func (m MovieModel) Delete(id int) error {
	return nil
}

type Movie struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int       `json:"year,omitzero"`
	Runtime   Runtime   `json:"runtime,omitzero"` // movie runtime in minutes
	Genres    []string  `json:"genres,omitzero"`
	Version   int       `json:"version"` // starts at 1 and increments each
	// time the movie information is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= time.Now().Year(), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

package data

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Define a User struct to represent an individual user. We are using json:"-"
// struct tag to prevent the Password and Version fields from appearing in any
// output when we encode it to JSON. Password field uses the custom password
// type defined below.
type User struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// Create a custom password type which is a struct containing the plaintext and
// hashed versions of the password for a user. The plaintext field is a *pointer*
// to a string, so that we're able to distinguish between a plaintext password
// not being present in the struct at all, versus a plaintext password which is
// the empty string "".
type password struct {
	plaintext *string
	hash      []byte
}

// Set() calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches() checks whether the provided plaintext password matches the hashed
// password stored in the struct, returning true if it matches and false if not.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

package main

import (
	"context"
	"net/http"

	"greenlight.cnoua.org/internal/data"
)

// Define a custom contextKey type, with the underlying type string.
type contextKey string

// Convert the string "user" to a contextKey type and assign it to the
// userContextKey constant. We'll use this constant as the key for getting and
// setting user information in the request context.
const userContextKey = contextKey("user")

// contextSetUser() returns a new copy of the request with the provided User
// struct added to the context. Using userContextKey constant as the key.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser() retrieves the User struct from the request context. The
// only time that we'll use this helper is when we logically expect there to be
// a User struct value in the context, and if it doesn't exist it will firmly
// be a programmer error. It's OK to PANIC in those circumstances.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}

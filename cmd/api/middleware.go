package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"

	"greenlight.cnoua.org/internal/data"
	"greenlight.cnoua.org/internal/validator"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create a deferred fn which will always be run in the event of a panic.
		defer func() {
			// use the built-in recover() fn to check if a panic occurred. If so, it
			// will return the panic value. If not, it will return nil.
			pv := recover()
			if pv != nil {
				// if there was a panic, set a "Connection: close" header on the response.
				// This acts as a trigger to make Go's HTTP server automatically close
				// the current connection after the response has ben sent.
				w.Header().Set("Connection", "close")
				// the value returned by recover() has the type any, so we use fmt.Errorf()
				// with the %v verb to coerce it into an error and call out
				// serverErrorResponse() helper. This will log the error at the ERROR
				// level and send the client a 500 Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%v", pv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// If rate limiting is not enabled, return the next handler in the chain with
	// no further action
	if !app.config.limiter.enabled {
		return next
	}
	// Define a client struct to hold the rate limiter & last seen for each client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	// Declare a mutex and a map to hold the clients IPs and rate limiters.
	var (
		mu sync.Mutex
		// Update the map so the values are pointers to a client struct.
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients
	// map once every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last 3 min,
			// delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	// The function we are returning is a closure, which 'closes over' the limiter
	// variable.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the realip.FromRequest() function to get the client's IP address.
		ip := realip.FromRequest(r)

		// Lock the mutex to prevent this code from being executed concurrently.
		mu.Lock()

		// Check to see if the IP address already exists in the map. If it doesn't,
		// initialize a new rate limiter and add the IP and limiter to the map.
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				// Use the requests-per-second & burst values from config struct.
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}

		// Update the last seen time for the client.
		clients[ip].lastSeen = time.Now()

		// Call the Allow() on the rate limiter for the current IP. If the Request
		// isn't permitted, unlock the mutex and send a 429 Too Many Requests response.
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		// Very important to unlock the mutex before calling the next handler in the
		// chain. Notice we don't use defer to unlock the mutex, as that would mean
		// the mutex isn't unlocked until all the handlers downstream of this
		// middleware have also returned.
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to
		// any caches that the response may vary based on the value of the
		// Authorization header in the request.
		w.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request. This will
		// return the empty string "" if there is no such header found.
		authorizationHeader := r.Header.Get("Authorization")

		// If there is no Authorization header found, use the contextSetUser() helper
		// to add the AnonymousUser to the request context. Then we call the next
		// handler in the chain and return without executing any of the code below.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, we expect the value of the Authorization header to be in the
		// format "Bearer <token>". We try to split this into its constituent parts,
		// and if the header isn't in the expected format we return a 401 Unauthorized
		// response using the invalidAuthenticationTokenResponse() helper.
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Extract the actual authentication token from the header parts.
		token := headerParts[1]

		// Validate the token to make sure it is in a sensible format.
		v := validator.New()

		// If the token isn't valid, use the invalidAuthenticationTokenResponse()
		// helper to send a response, rather than the failedValidationResponse()
		// helper that we'd normally use.
		if data.ValidateTokenPlainText(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Retrieve the details of the user associated with the authentication token,
		// again, calling invalidAuthenticationTokenResponse() if no matching record
		// was found. NOTICE that we are using ScopeAuthentication as first parameter.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Call the contextSetUser() helper to add the user information to the
		// request context.
		r = app.contextSetUser(r, user)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

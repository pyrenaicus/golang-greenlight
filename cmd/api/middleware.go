package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
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
	// Declare a mutex and a map to hold the clients IPs and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)

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
			clients[ip] = rate.NewLimiter(2, 4)
		}
		// Call the Allow() on the rate limiter for the current IP. If the Request
		// isn't permitted, unlock the mutex and send a 429 Too Many Requests response.
		if !clients[ip].Allow() {
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

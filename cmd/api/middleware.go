package main

import (
	"fmt"
	"net/http"
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

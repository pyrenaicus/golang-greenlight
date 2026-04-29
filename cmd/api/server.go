package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (app *application) serve() error {
	// Declare an HTTP server using the same settings as in our main() function.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Log a "starting server" message.
	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// Start the server as normal, returning any error.
	return srv.ListenAndServe()
}

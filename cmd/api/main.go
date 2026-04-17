package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	// we alias this import to the blank identifier to stop the compiler
	// complaining that the package isn't being used.
	_ "github.com/lib/pq"
)

const version = "1.0.0"

// a struct holding all the configuration settings for our app.
// network port & current operating environment.
type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

// a struct holding the dependencies for our HTTP handelrs, helpers
// and middleware.
type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// read DSN value from the db-dsn command-line flag into the config struct,
	// default to our development DSN if no flag is provided.
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://greenlight:pa55word@localhost/greenlight", "PostgreSQL DSN")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// call openDB() helper fn to create a connection pool, passing tje config
	// struct as an argument. If this returns an error, log it and exit.
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// defer a call to db.Close() so the connection pool is closed before the
	// main() fn exits.
	defer db.Close()

	// log a success message.
	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
	}

	// declare a HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// start the HTTP server
	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB() returns a sql.DB connection pool.
func openDB(cfg config) (*sql.DB, error) {
	// use sql.Open() to create an empty connection pool using the DSN from config.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// use PingContext() to establish a new connection to the db, passing the
	// context as a parameter. If the connection cannot be established within
	// 5 seconds, this will return an error, then close connection pool and
	// return the error.
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// return the sql.DB connection pool.
	return db, nil
}

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

// a struct holding all the configuration settings for our app.
// network port & current operating environment.
type config struct {
	port	int
	env		string
}

// a struct holding the dependencies for our HTTP handelrs, helpers
// and middleware.
type application struct {
	config	config
	logger	*slog.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		config:	cfg,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.
}

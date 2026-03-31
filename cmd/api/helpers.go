package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t") // no line prefix & tab indents
	if err != nil {
		return err
	}

	js = append(js, '\n')

	// loop through headers map, which behind the scenes has the type
	// map[string][]string, and add all the header keys & values to
	// the http.ResponseWriter's header map.
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// decode the request body into the target destination.
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// if error has type *json.SyntaxError return error message including
		// the location of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// in some circumstances with JSON syntax errors Decode() may return
		// an io.ErrUnexpectedEOF error, check for this using errors.Is() and
		// return a generic error message.
		// See: https://github.com/golang/go/issues/25956
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// json.UnmarshalTypeError occur when the JSON value is the wrong type for
		// the target destination. If the error relates to a specific field, we
		// include it in the error message.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// Decode() will return an io.EOF error if the request body is empty.
		// Check with errors.IS() and return a plain-english error message.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// a json.InvalidUnmarshalError will be returnes if we pass something that is
		// not a non-nil pointer as the target destination to Decode(). If this
		// happens we panic, rather than returning an error to our handler.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// for any other error, return it as-is.
		default:
			return err
		}
	}
	return nil
}

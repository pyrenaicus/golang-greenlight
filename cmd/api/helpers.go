package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	// limit the size of the request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	// initialize json.Decoder and call DisallowUnknownFields() method on it
	// before decoding. If the JSON includes any field that cannot be mapped
	// to the target destination, the decoder will return an error instead
	// of ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// decode the request body to the destination
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

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

		// if JSON contains a field which cannot be mapped to target destination,
		// Decode() will return an error message in the format "json: unknown
		// field "<name>"". We check for this, extract the field name from the
		// error, and interpolate it into our custom error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// check if error has the type *http.MaxBytesError, if it does, return
		// a clear error messagge.
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

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

	// call Decode() again, using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a single JSON value
	// this will return an io.EOF error. If we get anything else, we know that
	// there is additional data and we return custom error message.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// readString() returns a string value from the query string, or the provided
// default value if no matching key could be found.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// extract the value for a given key from the query string. If no key exists
	// it will return the empty string.
	s := qs.Get(key)

	// If no key exists (or the value is emmpty) return default value.
	if s == "" {
		return defaultValue
	}

	// Otherwise return the string.
	return s
}

package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// define an error that UnmarshalJSON() can return if we're unable to parse
// or convert the JSON string successfully.
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

// implement UnmarshalJSON() on Runtime type so that it satisfies
// json.Unmarshaler interface. Because UnmarshalJSON() needs to modify
// the receiver, we must use a pointer receiver.
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// we expect the incoming JSON value will be a string in the format
	// "<runtime> mins", first thing, we need to remove the surrounding
	// double quotes. If we can't, we return ErrInvalidRuntimeFormat
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// split the string to isolate the number
	parts := strings.Split(unquotedJSONValue, " ")

	// sanity check to make sure is in the expected format.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// parse the string into an int.
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// convert the int to a Runtime type and assign it to the receiver.
	*r = Runtime(i)

	return nil
}

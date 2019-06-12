package jsonapi

import (
	"io"
	"reflect"

	"github.com/google/jsonapi"
)

// UnmarshalPayload a request
func UnmarshalPayload(in io.Reader, model interface{}) error {
	return jsonapi.UnmarshalPayload(in, model)
}

// UnmarshalManyPayload a request
func UnmarshalManyPayload(in io.Reader, t reflect.Type) ([]interface{}, error) {
	return jsonapi.UnmarshalManyPayload(in, t)
}

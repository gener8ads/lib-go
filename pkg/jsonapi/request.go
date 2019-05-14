package jsonapi

import (
	"io"

	"github.com/google/jsonapi"
)

// UnmarshalPayload a request
func UnmarshalPayload(in io.Reader, model interface{}) error {
	return jsonapi.UnmarshalPayload(in, model)
}

package jsonapi

import (
	"github.com/google/jsonapi"
)

// Marshal a response
func Marshal(models interface{}) (jsonapi.Payloader, error) {
	return jsonapi.Marshal(models)
}

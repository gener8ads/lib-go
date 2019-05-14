package jsonapi

import (
	"github.com/google/jsonapi"
)

// Marshall a response
func Marshal(models interface{}) (jsonapi.Payloader, error) {
	return jsonapi.Marshal(models)
}

package jsonapi

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
)

// Marshal a response
func Marshal(models interface{}) (jsonapi.Payloader, error) {
	return jsonapi.Marshal(models)
}

// ErrorResponse params
type ErrorResponse struct {
	Status  int
	Code    string
	Pointer string
	Detail  string
}

type jsonAPIError struct {
	Errors []errorObject `json:"errors"`
}

// See https://jsonapi.org/format/#error-objects for more info
type errorObject struct {
	ID     string                  `json:"id,omitempty"`
	Title  string                  `json:"title,omitempty"`
	Detail string                  `json:"detail,omitempty"`
	Status string                  `json:"status,omitempty"`
	Code   string                  `json:"code,omitempty"`
	Source errorSource             `json:"source,omitempty"`
	Meta   *map[string]interface{} `json:"meta,omitempty"`
}

type errorSource struct {
	Pointer string `json:"pointer,omitempty"`
}

// Error responder - we use this instead of jsonapi.MarshalErrors to use the `Source` fields
func Error(c *gin.Context, err ErrorResponse) {
	pointer := err.Pointer

	if pointer == "" {
		pointer = "data/"
	} else if !strings.Contains(pointer, "attribute") {
		pointer = "data/attribute/" + pointer
	}

	code := err.Code

	if !strings.HasPrefix(code, "errors.") {
		code = "errors." + code
	}

	res := jsonAPIError{
		Errors: []errorObject{
			errorObject{
				Code:   code,
				Status: strconv.Itoa(err.Status),
				Detail: err.Detail,
				Source: errorSource{
					Pointer: pointer,
				},
			},
		},
	}

	c.AbortWithStatusJSON(err.Status, res)
}

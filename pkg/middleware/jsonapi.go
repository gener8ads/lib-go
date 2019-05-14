package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
)

// Can be removed when https://github.com/google/jsonapi/pull/134 is merged
type apiError struct {
	jsonapi.ErrorObject

	// Source is used to indicate which part of the request document caused the error.
	Source *apiErrorSource `json:"source,omitempty"`
}

type apiErrorSource struct {
	// Pointer is a JSON Pointer [RFC6901] to the associated entity in the request document [e.g. "/data" for a primary data object, or "/data/attributes/title" for a specific attribute].
	Pointer string `json:"pointer,omitempty"`

	// Parameter is a string indicating which URI query parameter caused the error.
	Parameter string `json:"parameter,omitempty"`
}

// JSONAPIContentType adds the correct JSON API content type
func JSONAPIContentType(c *gin.Context) {
	c.Header("Content-Type", jsonapi.MediaType)
}

// JSONAPIError takes the errors from the gin context and formats them as JSON API errors
func JSONAPIError(c *gin.Context) {
	c.Next()

	if len(c.Errors) == 0 {
		return
	}

	errors := []*apiError{}

	for _, e := range c.Errors {
		err := apiError{
			Source: &apiErrorSource{
				Pointer: "/data",
			},
		}
		err.Detail = e.Error()
		errors = append(errors, &err)
	}

	c.JSON(-1, gin.H{
		"errors": errors,
	})
}

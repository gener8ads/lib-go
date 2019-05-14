package jsonapi

import (
	"math"
	"net/url"
	"strconv"

	"github.com/google/jsonapi"
)

// Pagination takes an array and converts to JSON API struct
type Pagination struct {
	CurrentURL  *url.URL
	Data        interface{}
	Count       int
	CurrentPage int
	PerPage     int
}

// Marshal returns a JSON API shaped struct
func (t Pagination) Marshal() (payload *jsonapi.ManyPayload, err error) {
	marshalled, transformErr := jsonapi.Marshal(t.Data)

	if transformErr != nil {
		return payload, transformErr
	}

	payload = marshalled.(*jsonapi.ManyPayload)
	payload.Links = t.links()
	payload.Meta = t.meta()

	return payload, nil
}

func (t Pagination) links() *jsonapi.Links {
	firstURL := *t.CurrentURL
	firstQ := firstURL.Query()
	firstQ.Set("page", "1")
	firstURL.RawQuery = firstQ.Encode()

	lastURL := *t.CurrentURL
	lastQ := lastURL.Query()
	lastQ.Set("page", strconv.Itoa(t.pageCount()))
	lastURL.RawQuery = lastQ.Encode()

	nextURL := *t.CurrentURL
	nextQ := nextURL.Query()
	nextQ.Set("page", strconv.Itoa(t.CurrentPage+1))
	nextURL.RawQuery = nextQ.Encode()

	prevURL := *t.CurrentURL
	prevQ := prevURL.Query()
	prevPage := math.Max(1, float64(t.CurrentPage-1))
	prevQ.Set("page", strconv.Itoa(int(prevPage)))
	prevURL.RawQuery = prevQ.Encode()

	return &jsonapi.Links{
		"self":  t.CurrentURL.String(),
		"first": firstURL.String(),
		"last":  lastURL.String(),
		"next":  nextURL.String(),
		"prev":  prevURL.String(),
	}
}

func (t Pagination) meta() *jsonapi.Meta {
	return &jsonapi.Meta{
		"count":       t.Count,
		"total-pages": t.pageCount(),
	}
}

func (t Pagination) pageCount() int {
	perPage := float64(t.PerPage)
	count := float64(t.Count)

	pages := math.Ceil(count / perPage)

	return int(pages)
}

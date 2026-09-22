package v2

import (
	"context"
	"encoding/json"
	"fmt"
)

// pagedCollectionBody matches the {"results": [...], "__next": "..."} shape
// SAP's OData V2 services return for a collection response, with
// server-driven paging: "__next", when present, is a complete absolute URL
// for the next page and must be followed as-is rather than reconstructed.
type pagedCollectionBody struct {
	Results json.RawMessage `json:"results"`
	Next    string          `json:"__next"`
}

// DecodeCollectionPage unmarshals one page of a collection response's
// "d.results" into v (a pointer to a slice) and returns the "d.__next" link
// if SAP indicated more pages follow. An empty nextLink means v holds the
// last (or only) page.
func DecodeCollectionPage(body []byte, v interface{}) (nextLink string, err error) {
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return "", fmt.Errorf("odata: decoding envelope: %w", err)
	}

	var page pagedCollectionBody
	if err := json.Unmarshal(env.D, &page); err != nil {
		return "", fmt.Errorf("odata: decoding collection page: %w", err)
	}
	if err := json.Unmarshal(page.Results, v); err != nil {
		return "", fmt.Errorf("odata: decoding collection results: %w", err)
	}
	return page.Next, nil
}

// GetAllPages fetches every page of the collection at path, starting from
// the given base client, decoding each page into a []T and appending them
// into a single slice, following SAP's "__next" server-driven paging link
// until the server stops returning one. Callers must not silently assume a
// single-page response for any entity set SAP documents as potentially
// large — Partner Directory's StringParameters and BinaryParameters in
// particular can hold many entries per tenant.
func GetAllPages[T any](ctx context.Context, c *Client, path string) ([]T, error) {
	var all []T
	next := path

	for next != "" {
		body, err := c.Get(ctx, next)
		if err != nil {
			return nil, err
		}

		var page []T
		nextLink, err := DecodeCollectionPage(body, &page)
		if err != nil {
			return nil, err
		}

		all = append(all, page...)
		next = nextLink
	}

	return all, nil
}

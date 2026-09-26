package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Envelope is the OData V2 JSON response wrapper: a single entity is
// returned as {"d": {...}}, a collection as {"d": {"results": [...]}}.
type Envelope struct {
	D json.RawMessage `json:"d"`
}

// collectionBody matches the {"results": [...]} shape nested inside "d" for
// collection responses.
type collectionBody struct {
	Results json.RawMessage `json:"results"`
}

// DecodeEntity unmarshals a single-entity OData V2 response body into v.
func DecodeEntity(body []byte, v interface{}) error {
	if EmptyBody(body) {
		return ErrEmptyBody
	}
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("odata: decoding envelope: %w", err)
	}
	if err := json.Unmarshal(env.D, v); err != nil {
		return fmt.Errorf("odata: decoding entity: %w", err)
	}
	return nil
}

// DecodeCollection unmarshals a collection OData V2 response body's
// "d.results" array into v, which must be a pointer to a slice.
func DecodeCollection(body []byte, v interface{}) error {
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("odata: decoding envelope: %w", err)
	}

	var coll collectionBody
	if err := json.Unmarshal(env.D, &coll); err != nil {
		return fmt.Errorf("odata: decoding collection: %w", err)
	}
	if err := json.Unmarshal(coll.Results, v); err != nil {
		return fmt.Errorf("odata: decoding collection results: %w", err)
	}
	return nil
}

// ErrEmptyBody reports a successful response without an entity. SAP's
// Security Content and Message Store services answer writes with
// 202 Accepted and an empty body, so a caller that needs the entity has to
// read it back.
var ErrEmptyBody = errors.New("odata: SAP accepted the request but returned no entity in the response")

// EmptyBody reports whether a response body carries no content.
func EmptyBody(body []byte) bool {
	return len(bytes.TrimSpace(body)) == 0
}

// Package apierror defines the typed error shared by every SAP Integration
// Suite API client (OData V2, OData V4, and any plain REST client), so the
// provider layer can turn a failure into a Terraform diagnostic without
// caring which wire format produced it.
package apierror

import "fmt"

// Detail is one entry of a SAP API error's detail list.
type Detail struct {
	Code    string
	Message string
}

// Error is a normalized SAP API error.
type Error struct {
	// StatusCode is the HTTP status code the API returned.
	StatusCode int
	// Code is the SAP-specific error code, if the API supplied one.
	Code string
	// Message is the human-readable error message, if the API supplied one.
	Message string
	// Details holds any additional error entries the API supplied.
	Details []Detail
	// RequestID is a correlation ID for the failed request, if present.
	RequestID string
}

func (e *Error) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "no error message returned by the API"
	}
	if e.Code != "" {
		return fmt.Sprintf("SAP API returned HTTP %d, code %q: %s", e.StatusCode, e.Code, msg)
	}
	return fmt.Sprintf("SAP API returned HTTP %d: %s", e.StatusCode, msg)
}

// IsNotFound reports whether the error represents an HTTP 404, which every
// resource in this provider treats as "already gone" on delete and "removed
// out of band" on read.
func (e *Error) IsNotFound() bool {
	return e != nil && e.StatusCode == 404
}

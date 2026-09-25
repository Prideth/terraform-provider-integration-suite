package v2

import (
	"fmt"
	"regexp"
	"strings"
)

// runtimeLocationIDPattern limits runtime location IDs to characters that are
// safe in a URL path segment without escaping. SAP shows these IDs in the
// Integration Suite monitoring URL, for example {"runtimeLocationId":"myedge"}.
var runtimeLocationIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// ValidateRuntimeLocationID reports whether id can be used as a runtime
// location segment.
func ValidateRuntimeLocationID(id string) error {
	if !runtimeLocationIDPattern.MatchString(id) {
		return fmt.Errorf("odata: %q is not a valid runtime location ID (letters, digits, '.', '_' and '-' only)", id)
	}
	return nil
}

// ServiceRoot returns the Cloud Integration API service root for host. With an
// empty runtimeLocationID it is host + "/api/v1" (the cloud runtime); otherwise
// it is host + "/location/<runtimeLocationID>/api/v1", the form SAP documents
// for addressing an Edge Integration Cell through the same APIs.
func ServiceRoot(host, runtimeLocationID string) (string, error) {
	root := strings.TrimRight(host, "/")
	if runtimeLocationID == "" {
		return root + "/api/v1", nil
	}
	if err := ValidateRuntimeLocationID(runtimeLocationID); err != nil {
		return "", err
	}
	return root + "/location/" + runtimeLocationID + "/api/v1", nil
}

// WithBaseURL returns a copy of c that sends requests to baseURL, sharing the
// same HTTP transport.
func (c *Client) WithBaseURL(baseURL string) *Client {
	clone := *c
	clone.baseURL = baseURL
	return &clone
}

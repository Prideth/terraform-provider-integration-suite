package apicomposition

import (
	"encoding/json"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
)

// errorBody matches the OData V4 standard error response format
// (message as a plain string, unlike OData V2's nested {"lang","value"}
// object): {"error": {"code": "...", "message": "..."}}. This project did
// not capture a verbatim error response example for this API during
// research — SAP's documentation only showed success responses — so this
// shape is inferred from the OData V4 specification itself rather than
// confirmed against a worked SAP example. If body does not parse this way,
// the resulting error still carries statusCode so callers can act on
// IsNotFound and similar checks.
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ParseError builds an apierror.Error from an API Composition Configuration
// API error response body.
func ParseError(statusCode int, body []byte) *apierror.Error {
	result := &apierror.Error{StatusCode: statusCode}

	var parsed errorBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return result
	}

	result.Code = parsed.Error.Code
	result.Message = parsed.Error.Message
	return result
}

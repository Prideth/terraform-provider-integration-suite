package v2

import (
	"encoding/json"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
)

// errorBody matches the OData V2 error response format:
//
//	{"error": {"code": "...", "message": {"lang": "en-US", "value": "..."}, "innererror": {...}}}
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message struct {
			Lang  string `json:"lang"`
			Value string `json:"value"`
		} `json:"message"`
		InnerError struct {
			Errordetails []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"errordetails"`
		} `json:"innererror"`
	} `json:"error"`
}

// ParseError builds an apierror.Error from an OData V2 error response body.
// If body does not parse as an OData V2 error (for example, an upstream
// gateway returned plain text), the resulting error still carries
// statusCode so callers can act on IsNotFound and similar checks.
func ParseError(statusCode int, body []byte) *apierror.Error {
	result := &apierror.Error{StatusCode: statusCode}

	var parsed errorBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return result
	}

	result.Code = parsed.Error.Code
	result.Message = parsed.Error.Message.Value

	for _, d := range parsed.Error.InnerError.Errordetails {
		result.Details = append(result.Details, apierror.Detail{
			Code:    d.Code,
			Message: d.Message,
		})
	}

	return result
}

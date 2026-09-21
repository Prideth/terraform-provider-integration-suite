package v2

import "testing"

func TestParseError(t *testing.T) {
	body := []byte(`{
		"error": {
			"code": "BAD_REQUEST",
			"message": {"lang": "en-US", "value": "The package ID is invalid"},
			"innererror": {
				"errordetails": [
					{"code": "FIELD_REQUIRED", "message": "Name is required"}
				]
			}
		}
	}`)

	err := ParseError(400, body)

	if err.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", err.StatusCode)
	}
	if err.Code != "BAD_REQUEST" {
		t.Errorf("Code = %q, want BAD_REQUEST", err.Code)
	}
	if err.Message != "The package ID is invalid" {
		t.Errorf("Message = %q, want %q", err.Message, "The package ID is invalid")
	}
	if len(err.Details) != 1 || err.Details[0].Code != "FIELD_REQUIRED" {
		t.Errorf("Details = %+v", err.Details)
	}
}

func TestParseError_NonJSONBody(t *testing.T) {
	err := ParseError(503, []byte("<html>Service Unavailable</html>"))
	if err.StatusCode != 503 {
		t.Errorf("StatusCode = %d, want 503", err.StatusCode)
	}
	if err.Code != "" || err.Message != "" {
		t.Errorf("expected empty Code/Message for a non-JSON body, got %+v", err)
	}
}

func TestParseError_IsNotFound(t *testing.T) {
	err := ParseError(404, []byte(`{}`))
	if !err.IsNotFound() {
		t.Error("expected IsNotFound() to be true for a 404")
	}

	err = ParseError(400, []byte(`{}`))
	if err.IsNotFound() {
		t.Error("expected IsNotFound() to be false for a 400")
	}
}

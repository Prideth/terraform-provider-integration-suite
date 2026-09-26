package cloudintegration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateNumberRange_ExactRequestBody(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	initial := "0"
	err := client.CreateNumberRange(context.Background(), NumberRange{
		Name:         "My NRO Object",
		MinValue:     "0",
		MaxValue:     "9999",
		Description:  " Number Range Object ",
		Rotate:       true,
		FieldLength:  "4",
		CurrentValue: &initial,
	})
	if err != nil {
		t.Fatalf("CreateNumberRange() error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/v1/NumberRanges" {
		t.Errorf("path = %q, want /api/v1/NumberRanges", gotPath)
	}

	// SAP's own documented example, verbatim.
	want := `{"CurrentValue":"0","Name":"My NRO Object","MinValue":"0","MaxValue":"9999","Description":" Number Range Object ","Rotate":"true","FieldLength":"4"}`
	if string(gotBody) != want {
		t.Errorf("request body = %s, want %s", gotBody, want)
	}
}

func TestClient_CreateNumberRange_RequiresCurrentValue(t *testing.T) {
	client := New(http.DefaultClient, "https://example.invalid")

	err := client.CreateNumberRange(context.Background(), NumberRange{
		Name:     "My NRO Object",
		MinValue: "0",
		MaxValue: "9999",
	})
	if err == nil {
		t.Fatal("CreateNumberRange() error = nil, want an error when CurrentValue is nil")
	}
}

func TestClient_CreateNumberRange_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error": {"code": "409", "message": {"value": "Conflict"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	current := "0"
	err := client.CreateNumberRange(context.Background(), NumberRange{Name: "Dup", CurrentValue: &current})
	if err == nil {
		t.Fatal("CreateNumberRange() error = nil, want an error for a 409 response")
	}
}

func TestClient_CreateNumberRange_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "403", "message": {"value": "Forbidden"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	current := "0"
	err := client.CreateNumberRange(context.Background(), NumberRange{Name: "X", CurrentValue: &current})
	if err == nil {
		t.Fatal("CreateNumberRange() error = nil, want an error for a 403 response")
	}
}

func TestClient_CreateNumberRange_MalformedErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	current := "0"
	err := client.CreateNumberRange(context.Background(), NumberRange{Name: "X", CurrentValue: &current})
	if err == nil {
		t.Fatal("CreateNumberRange() error = nil, want an error for a malformed 400 response")
	}
}

func TestClient_UpdateNumberRange_ExactRequestBody(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	current := "0"
	err := client.UpdateNumberRange(context.Background(), NumberRange{
		Name:         "My NRO Object",
		MinValue:     "0",
		MaxValue:     "9999",
		Description:  " Number Range Object ",
		Rotate:       true,
		FieldLength:  "4",
		CurrentValue: &current,
	})
	if err != nil {
		t.Fatalf("UpdateNumberRange() error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if gotPath != "/api/v1/NumberRanges('My NRO Object')" {
		t.Errorf("path = %q, want /api/v1/NumberRanges('My NRO Object')", gotPath)
	}

	want := `{"CurrentValue":"0","Name":"My NRO Object","MinValue":"0","MaxValue":"9999","Description":" Number Range Object ","Rotate":"true","FieldLength":"4"}`
	if string(gotBody) != want {
		t.Errorf("request body = %s, want %s", gotBody, want)
	}
}

// TestClient_UpdateNumberRange_RejectsMissingCurrentValue: a tenant answered
// a PUT without CurrentValue with 500 and left the object unchanged
// (September 2026). The client refuses such an update before sending it.
func TestClient_UpdateNumberRange_RejectsMissingCurrentValue(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	err := client.UpdateNumberRange(context.Background(), NumberRange{
		Name: "MyRange", MinValue: "0", MaxValue: "9999", Rotate: true, FieldLength: "4",
	})
	if err == nil {
		t.Fatal("expected an error for an update without CurrentValue")
	}
	if called {
		t.Error("the update was sent although SAP rejects it")
	}
}

func TestClient_UpdateNumberRange_EscapesNameInPath(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	zero := "0"
	err := client.UpdateNumberRange(context.Background(), NumberRange{Name: "Customer's Range", CurrentValue: &zero})
	if err != nil {
		t.Fatalf("UpdateNumberRange() error: %v", err)
	}

	want := "/api/v1/NumberRanges('Customer''s Range')"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestClient_UpdateNumberRange_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "Not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.UpdateNumberRange(context.Background(), NumberRange{Name: "Missing"})
	if err == nil {
		t.Fatal("UpdateNumberRange() error = nil, want an error for a 404 response")
	}
}

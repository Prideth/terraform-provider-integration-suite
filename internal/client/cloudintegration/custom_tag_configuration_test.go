package cloudintegration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetCustomTagConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/CustomTagConfigurations('CustomTags')/$value"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		// $value returns the raw configuration JSON, never wrapped in the
		// standard OData "d" envelope.
		_, _ = w.Write([]byte(`{"customTagsConfiguration":[{"tagName":"Author","permittedValues":["Mr. Bean","Ms. Bean"],"isMandatory":true}]}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	tags, err := client.GetCustomTagConfiguration(context.Background())
	if err != nil {
		t.Fatalf("GetCustomTagConfiguration() error: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("len(tags) = %d, want 1", len(tags))
	}
	if tags[0].Name != "Author" {
		t.Errorf("Name = %q, want Author", tags[0].Name)
	}
	if !tags[0].Mandatory {
		t.Error("Mandatory = false, want true")
	}
	if len(tags[0].PermittedValues) != 2 {
		t.Errorf("PermittedValues = %v, want 2 entries", tags[0].PermittedValues)
	}
}

func TestClient_GetCustomTagConfiguration_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "Not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetCustomTagConfiguration(context.Background()); err == nil {
		t.Fatal("GetCustomTagConfiguration() error = nil, want an error for a 404 response (no configuration exists yet)")
	}
}

func TestClient_GetCustomTagConfiguration_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetCustomTagConfiguration(context.Background()); err == nil {
		t.Fatal("GetCustomTagConfiguration() error = nil, want a decoding error for a malformed response")
	}
}

// TestClient_SetCustomTagConfiguration_ExactRequestBody verifies the exact
// documented request shape: POST to CustomTagConfigurations with
// Overwrite=true, a JSON body of {"CustomTagsConfigurationContent":
// "<base64>"}, decoding to exactly SAP's own documented example content.
func TestClient_SetCustomTagConfiguration_ExactRequestBody(t *testing.T) {
	var gotPath, gotQuery string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.SetCustomTagConfiguration(context.Background(), []CustomTag{
		{Name: "Owner", Mandatory: true},
	})
	if err != nil {
		t.Fatalf("SetCustomTagConfiguration() error: %v", err)
	}

	if gotPath != "/api/v1/CustomTagConfigurations" {
		t.Errorf("path = %q, want /api/v1/CustomTagConfigurations", gotPath)
	}
	if gotQuery != "Overwrite=true" {
		t.Errorf("query = %q, want Overwrite=true", gotQuery)
	}

	var envelope customTagConfigurationWriteRequest
	if err := json.Unmarshal(gotBody, &envelope); err != nil {
		t.Fatalf("decoding request body: %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(envelope.CustomTagsConfigurationContent)
	if err != nil {
		t.Fatalf("decoding base64 content: %v", err)
	}

	// SAP's own documented example, verbatim: creating a mandatory custom
	// tag named "Owner" produces exactly this decoded JSON.
	want := `{"customTagsConfiguration":[{"tagName":"Owner","isMandatory":true}]}`
	if string(decoded) != want {
		t.Errorf("decoded content = %s, want %s", decoded, want)
	}
}

// TestClient_SetCustomTagConfiguration_DeterministicOrdering proves the
// request body does not depend on the order tags or permitted values were
// supplied in: two calls with the same tags in a different order produce
// byte-identical request bodies.
func TestClient_SetCustomTagConfiguration_DeterministicOrdering(t *testing.T) {
	var bodies [][]byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		bodies = append(bodies, body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	orderA := []CustomTag{
		{Name: "Owner", Mandatory: true, PermittedValues: []string{"Bravo", "Alpha"}},
		{Name: "BusinessUnit", Mandatory: false, PermittedValues: []string{"Water", "Network"}},
	}
	orderB := []CustomTag{
		{Name: "BusinessUnit", Mandatory: false, PermittedValues: []string{"Network", "Water"}},
		{Name: "Owner", Mandatory: true, PermittedValues: []string{"Alpha", "Bravo"}},
	}

	if err := client.SetCustomTagConfiguration(context.Background(), orderA); err != nil {
		t.Fatalf("SetCustomTagConfiguration(orderA) error: %v", err)
	}
	if err := client.SetCustomTagConfiguration(context.Background(), orderB); err != nil {
		t.Fatalf("SetCustomTagConfiguration(orderB) error: %v", err)
	}

	if len(bodies) != 2 {
		t.Fatalf("got %d requests, want 2", len(bodies))
	}
	if string(bodies[0]) != string(bodies[1]) {
		t.Errorf("request bodies differ by input order:\n  A: %s\n  B: %s", bodies[0], bodies[1])
	}
}

func TestClient_SetCustomTagConfiguration_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error": {"code": "409", "message": {"value": "Conflict"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.SetCustomTagConfiguration(context.Background(), []CustomTag{{Name: "Owner", Mandatory: true}})
	if err == nil {
		t.Fatal("SetCustomTagConfiguration() error = nil, want an error for a 409 response")
	}
}

func TestClient_SetCustomTagConfiguration_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "403", "message": {"value": "Forbidden"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.SetCustomTagConfiguration(context.Background(), []CustomTag{{Name: "Owner", Mandatory: true}})
	if err == nil {
		t.Fatal("SetCustomTagConfiguration() error = nil, want an error for a 403 response")
	}
}

func TestSortedCustomTags(t *testing.T) {
	in := []CustomTag{
		{Name: "Zeta", PermittedValues: []string{"c", "a", "b"}},
		{Name: "Alpha", PermittedValues: []string{"z", "y"}},
	}
	got := sortedCustomTags(in)

	if got[0].Name != "Alpha" || got[1].Name != "Zeta" {
		t.Errorf("tags not sorted by name: %+v", got)
	}
	if got[1].PermittedValues[0] != "a" || got[1].PermittedValues[1] != "b" || got[1].PermittedValues[2] != "c" {
		t.Errorf("Zeta's permitted values not sorted: %v", got[1].PermittedValues)
	}
	if got[0].PermittedValues[0] != "y" || got[0].PermittedValues[1] != "z" {
		t.Errorf("Alpha's permitted values not sorted: %v", got[0].PermittedValues)
	}

	// The input slice itself must not be mutated.
	if in[0].Name != "Zeta" || in[0].PermittedValues[0] != "c" {
		t.Error("sortedCustomTags mutated its input")
	}
}

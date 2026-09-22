package cloudintegration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// syntheticESAContent is an unmistakably synthetic ".esa" fixture, not a
// real SAP or third-party adapter artifact. It is never a real archive;
// this client transports content opaquely and never parses it, so its
// actual bytes are irrelevant to every test below.
var syntheticESAContent = []byte("tf-acc-synthetic-esa-fixture-not-a-real-adapter")

func TestClient_CreateIntegrationAdapter(t *testing.T) {
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/IntegrationAdapterDesigntimeArtifacts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading POST body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Id": "custom-sftp-extension", "Name": "Custom SFTP Extension", "PackageId": "ADAPTERS", "Version": "1.0.0", "Type": "Analytics", "Application": "Slack"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	created, err := client.CreateIntegrationAdapter(context.Background(), IntegrationAdapter{
		ID:          "custom-sftp-extension",
		Name:        "Custom SFTP Extension",
		PackageID:   "ADAPTERS",
		Type:        "Analytics",
		Application: "Slack",
	}, syntheticESAContent)
	if err != nil {
		t.Fatalf("CreateIntegrationAdapter() error: %v", err)
	}
	if created.ID != "custom-sftp-extension" {
		t.Errorf("ID = %q, want custom-sftp-extension", created.ID)
	}
	if created.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", created.Version)
	}

	var decoded map[string]any
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("decoding POST body: %v", err)
	}
	if decoded["PackageId"] != "ADAPTERS" {
		t.Errorf("POST body PackageId = %v, want ADAPTERS", decoded["PackageId"])
	}
	if decoded["Type"] != "Analytics" {
		t.Errorf("POST body Type = %v, want Analytics", decoded["Type"])
	}
	content, ok := decoded["ArtifactContent"].(string)
	if !ok || content == "" {
		t.Errorf("POST body ArtifactContent = %v, want a non-empty base64 string", decoded["ArtifactContent"])
	}
}

func TestClient_CreateIntegrationAdapter_DuplicateID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error": {"code": "409", "message": {"value": "An integration adapter with this ID already exists"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.CreateIntegrationAdapter(context.Background(), IntegrationAdapter{
		ID: "custom-sftp-extension", Name: "Custom SFTP Extension", PackageID: "ADAPTERS",
	}, syntheticESAContent)
	if err == nil {
		t.Fatal("CreateIntegrationAdapter() error = nil, want an error for a duplicate ID (409)")
	}
}

func TestClient_GetIntegrationAdapter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/IntegrationAdapterDesigntimeArtifacts('custom-sftp-extension')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "custom-sftp-extension", "Name": "Custom SFTP Extension", "PackageId": "ADAPTERS", "Version": "1.0.0", "Type": "Analytics", "Application": "Slack"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	adapter, err := client.GetIntegrationAdapter(context.Background(), "custom-sftp-extension")
	if err != nil {
		t.Fatalf("GetIntegrationAdapter() error: %v", err)
	}
	if adapter.Name != "Custom SFTP Extension" {
		t.Errorf("Name = %q, want Custom SFTP Extension", adapter.Name)
	}
	if adapter.PackageID != "ADAPTERS" {
		t.Errorf("PackageId = %q, want ADAPTERS", adapter.PackageID)
	}
}

func TestClient_GetIntegrationAdapter_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "Not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetIntegrationAdapter(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("GetIntegrationAdapter() error = nil, want an error for a 404 response")
	}
}

func TestClient_DeleteIntegrationAdapter(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeleteIntegrationAdapter(context.Background(), "custom-sftp-extension"); err != nil {
		t.Fatalf("DeleteIntegrationAdapter() error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	want := "/api/v1/IntegrationAdapterDesigntimeArtifacts('custom-sftp-extension')"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestClient_DeleteIntegrationAdapter_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "Not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeleteIntegrationAdapter(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("DeleteIntegrationAdapter() error = nil, want an error for a 404 response")
	}
}

func TestClient_DeployIntegrationAdapter(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeployIntegrationAdapter(context.Background(), "custom-sftp-extension"); err != nil {
		t.Fatalf("DeployIntegrationAdapter() error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	want := "/api/v1/DeployIntegrationAdapterDesigntimeArtifact?Id='custom-sftp-extension'"
	if gotPath != want {
		t.Errorf("path = %q, want %q (singular \"Artifact\", no Version parameter)", gotPath, want)
	}
}

func TestClient_DeployIntegrationAdapter_DoesNotSendVersionParameter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "Version") {
			t.Errorf("deploy request must not include a Version query parameter, got %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeployIntegrationAdapter(context.Background(), "custom-sftp-extension"); err != nil {
		t.Fatalf("DeployIntegrationAdapter() error: %v", err)
	}
}

func TestClient_DeployIntegrationAdapter_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "403", "message": {"value": "Forbidden"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeployIntegrationAdapter(context.Background(), "custom-sftp-extension"); err == nil {
		t.Fatal("DeployIntegrationAdapter() error = nil, want an error for a 403 response")
	}
}

func TestClient_CreateIntegrationAdapter_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.CreateIntegrationAdapter(context.Background(), IntegrationAdapter{ID: "x", PackageID: "P"}, syntheticESAContent)
	if err == nil {
		t.Fatal("CreateIntegrationAdapter() error = nil, want a decoding error for a malformed response")
	}
}

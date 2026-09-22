package cloudintegration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
)

func TestClient_CreateMessageMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/MessageMappingDesigntimeArtifacts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var sent MessageMapping
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if sent.Content != base64.StdEncoding.EncodeToString([]byte("mapping-bytes")) {
			t.Errorf("Content was not base64-encoded correctly: %q", sent.Content)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Id": "customer-mapping", "Name": "Customer Mapping", "PackageId": "UTILITIES", "Version": "1.0.0"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	mapping, err := client.CreateMessageMapping(context.Background(), "UTILITIES", "customer-mapping", "Customer Mapping", []byte("mapping-bytes"))
	if err != nil {
		t.Fatalf("CreateMessageMapping() error: %v", err)
	}
	if mapping.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", mapping.Version)
	}
}

func TestClient_CreateMessageMapping_InvalidArtifactIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"code": "BAD_REQUEST", "message": {"value": "malformed message mapping archive"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.CreateMessageMapping(context.Background(), "UTILITIES", "broken-mapping", "Broken", []byte("not-a-zip"))

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected a 400 *apierror.Error, got %v", err)
	}
}

func TestClient_GetMessageMapping_UsesActiveVersionKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/MessageMappingDesigntimeArtifacts(Id='customer-mapping',Version='active')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "customer-mapping", "Name": "Customer Mapping", "PackageId": "UTILITIES", "Version": "1.0.0"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetMessageMapping(context.Background(), "UTILITIES", "customer-mapping"); err != nil {
		t.Fatalf("GetMessageMapping() error: %v", err)
	}
}

func TestClient_GetMessageMapping_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "no such message mapping"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.GetMessageMapping(context.Background(), "UTILITIES", "missing")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

func TestClient_GetMessageMapping_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "FORBIDDEN", "message": {"value": "missing WorkspacePackagesConfigure role"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.GetMessageMapping(context.Background(), "UTILITIES", "customer-mapping")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("expected a 403 *apierror.Error, got %v", err)
	}
}

func TestClient_GetMessageMapping_MalformedResponseIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetMessageMapping(context.Background(), "UTILITIES", "customer-mapping"); err == nil {
		t.Fatal("expected an error decoding a malformed OData response, got nil")
	}
}

func TestClient_UpdateMessageMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		want := "/api/v1/MessageMappingDesigntimeArtifacts(Id='customer-mapping',Version='active')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}

		var sent MessageMapping
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if sent.Content != base64.StdEncoding.EncodeToString([]byte("new-mapping-bytes")) {
			t.Errorf("Content was not base64-encoded correctly: %q", sent.Content)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "customer-mapping", "Name": "Customer Mapping v2", "PackageId": "UTILITIES", "Version": "1.0.1"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	mapping, err := client.UpdateMessageMapping(context.Background(), "customer-mapping", "Customer Mapping v2", []byte("new-mapping-bytes"))
	if err != nil {
		t.Fatalf("UpdateMessageMapping() error: %v", err)
	}
	if mapping.Version != "1.0.1" {
		t.Errorf("Version = %q, want 1.0.1", mapping.Version)
	}
}

func TestClient_UpdateMessageMapping_ConflictIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error": {"code": "CONFLICT", "message": {"value": "artifact is locked for editing"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.UpdateMessageMapping(context.Background(), "customer-mapping", "Customer Mapping v2", []byte("new-mapping-bytes"))

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusConflict {
		t.Fatalf("expected a 409 *apierror.Error, got %v", err)
	}
}

func TestClient_DeleteMessageMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/MessageMappingDesigntimeArtifacts(Id='customer-mapping',Version='active')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeleteMessageMapping(context.Background(), "customer-mapping"); err != nil {
		t.Fatalf("DeleteMessageMapping() error: %v", err)
	}
}

func TestClient_DeleteMessageMapping_NotFoundIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "gone"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeleteMessageMapping(context.Background(), "customer-mapping")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

func TestClient_DeployMessageMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		want := "/api/v1/DeployMessageMappingDesigntimeArtifact?Id='customer-mapping'&Version='1.0.1'"
		if r.URL.String() != want {
			t.Errorf("url = %q, want %q", r.URL.String(), want)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeployMessageMapping(context.Background(), "customer-mapping", "1.0.1"); err != nil {
		t.Fatalf("DeployMessageMapping() error: %v", err)
	}
}

func TestClient_DeployMessageMapping_NotFoundIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "no such message mapping"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeployMessageMapping(context.Background(), "missing", "active")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

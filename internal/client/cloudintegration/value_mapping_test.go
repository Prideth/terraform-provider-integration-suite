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

func TestClient_CreateValueMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ValueMappingDesigntimeArtifacts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var sent ValueMapping
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if sent.Content != base64.StdEncoding.EncodeToString([]byte("mapping-bytes")) {
			t.Errorf("Content was not base64-encoded correctly: %q", sent.Content)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Id": "company-codes", "Name": "Company Codes", "PackageId": "UTILITIES", "Version": "1.0.0"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	mapping, err := client.CreateValueMapping(context.Background(), "UTILITIES", "company-codes", "Company Codes", []byte("mapping-bytes"))
	if err != nil {
		t.Fatalf("CreateValueMapping() error: %v", err)
	}
	if mapping.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", mapping.Version)
	}
}

func TestClient_GetValueMapping_UsesActiveVersionKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/ValueMappingDesigntimeArtifacts(Id='company-codes',Version='active')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "company-codes", "Name": "Company Codes", "PackageId": "UTILITIES", "Version": "1.0.0"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetValueMapping(context.Background(), "UTILITIES", "company-codes"); err != nil {
		t.Fatalf("GetValueMapping() error: %v", err)
	}
}

func TestClient_GetValueMapping_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "no such value mapping"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.GetValueMapping(context.Background(), "UTILITIES", "missing")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

func TestClient_UpdateValueMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		want := "/api/v1/ValueMappingDesigntimeArtifacts(Id='company-codes',Version='active')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "company-codes", "Name": "Company Codes v2", "PackageId": "UTILITIES", "Version": "1.0.1"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	mapping, err := client.UpdateValueMapping(context.Background(), "company-codes", "Company Codes v2", []byte("new-mapping-bytes"))
	if err != nil {
		t.Fatalf("UpdateValueMapping() error: %v", err)
	}
	if mapping.Version != "1.0.1" {
		t.Errorf("Version = %q, want 1.0.1", mapping.Version)
	}
}

func TestClient_DeleteValueMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/ValueMappingDesigntimeArtifacts(Id='company-codes',Version='active')"
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

	if err := client.DeleteValueMapping(context.Background(), "company-codes"); err != nil {
		t.Fatalf("DeleteValueMapping() error: %v", err)
	}
}

func TestClient_DeleteValueMapping_NotFoundIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "gone"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeleteValueMapping(context.Background(), "company-codes")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

func TestClient_DeployValueMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		want := "/api/v1/DeployValueMappingDesigntimeArtifact?Id='company-codes'&Version='1.0.1'"
		if r.URL.String() != want {
			t.Errorf("url = %q, want %q", r.URL.String(), want)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeployValueMapping(context.Background(), "company-codes", "1.0.1"); err != nil {
		t.Fatalf("DeployValueMapping() error: %v", err)
	}
}

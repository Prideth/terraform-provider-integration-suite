package cloudintegration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
)

func TestClient_GetPackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/IntegrationPackages('UTILITIES')" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "UTILITIES", "Name": "Utilities", "Description": "Utilities package"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	pkg, err := client.GetPackage(context.Background(), "UTILITIES")
	if err != nil {
		t.Fatalf("GetPackage() error: %v", err)
	}
	if pkg.Name != "Utilities" {
		t.Errorf("Name = %q, want Utilities", pkg.Name)
	}
}

func TestClient_GetPackage_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "no such package"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.GetPackage(context.Background(), "MISSING")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

func TestClient_CreatePackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/IntegrationPackages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Id": "UTILITIES", "Name": "Utilities"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	pkg, err := client.CreatePackage(context.Background(), Package{ID: "UTILITIES", Name: "Utilities"})
	if err != nil {
		t.Fatalf("CreatePackage() error: %v", err)
	}
	if pkg.ID != "UTILITIES" {
		t.Errorf("ID = %q, want UTILITIES", pkg.ID)
	}
}

func TestClient_DeletePackage_NotFoundIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "gone"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeletePackage(context.Background(), "UTILITIES")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

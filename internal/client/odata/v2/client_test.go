package v2

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
)

func TestClient_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "UTILITIES"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	body, err := client.Get(context.Background(), "IntegrationPackages('UTILITIES')")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	var pkg testPackage
	if err := DecodeEntity(body, &pkg); err != nil {
		t.Fatalf("DecodeEntity() error: %v", err)
	}
	if pkg.ID != "UTILITIES" {
		t.Errorf("ID = %q, want UTILITIES", pkg.ID)
	}
}

func TestClient_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "no such package"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.Get(context.Background(), "IntegrationPackages('MISSING')")
	if err == nil {
		t.Fatal("expected an error for a 404 response")
	}

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *apierror.Error, got %T: %v", err, err)
	}
	if !apiErr.IsNotFound() {
		t.Errorf("expected IsNotFound() to be true, got status %d", apiErr.StatusCode)
	}
	if apiErr.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want NOT_FOUND", apiErr.Code)
	}
}

func TestClient_Post_SendsBody(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		receivedBody = string(buf)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Id": "UTILITIES"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.Post(context.Background(), "IntegrationPackages", []byte(`{"Id":"UTILITIES"}`))
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	if receivedBody != `{"Id":"UTILITIES"}` {
		t.Errorf("receivedBody = %q", receivedBody)
	}
}

func TestClient_Delete_TreatsNotFoundAsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.Delete(context.Background(), "IntegrationPackages('UTILITIES')"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
}

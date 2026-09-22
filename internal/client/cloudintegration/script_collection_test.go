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

func TestClient_CreateScriptCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ScriptCollectionDesigntimeArtifacts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var sent ScriptCollection
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if sent.Content != base64.StdEncoding.EncodeToString([]byte("script-bytes")) {
			t.Errorf("Content was not base64-encoded correctly: %q", sent.Content)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Id": "shared-scripts", "Name": "Shared Scripts", "PackageId": "UTILITIES", "Version": "1.0.0"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	sc, err := client.CreateScriptCollection(context.Background(), "UTILITIES", "shared-scripts", "Shared Scripts", []byte("script-bytes"))
	if err != nil {
		t.Fatalf("CreateScriptCollection() error: %v", err)
	}
	if sc.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", sc.Version)
	}
}

func TestClient_CreateScriptCollection_InvalidArtifactIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"code": "BAD_REQUEST", "message": {"value": "malformed script collection archive"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.CreateScriptCollection(context.Background(), "UTILITIES", "broken-scripts", "Broken", []byte("not-a-zip"))

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected a 400 *apierror.Error, got %v", err)
	}
}

func TestClient_GetScriptCollection_UsesActiveVersionKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/ScriptCollectionDesigntimeArtifacts(Id='shared-scripts',Version='active')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "shared-scripts", "Name": "Shared Scripts", "PackageId": "UTILITIES", "Version": "1.0.0"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetScriptCollection(context.Background(), "UTILITIES", "shared-scripts"); err != nil {
		t.Fatalf("GetScriptCollection() error: %v", err)
	}
}

func TestClient_GetScriptCollection_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "no such script collection"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.GetScriptCollection(context.Background(), "UTILITIES", "missing")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

func TestClient_GetScriptCollection_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "FORBIDDEN", "message": {"value": "missing WorkspacePackagesConfigure role"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.GetScriptCollection(context.Background(), "UTILITIES", "shared-scripts")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("expected a 403 *apierror.Error, got %v", err)
	}
}

func TestClient_GetScriptCollection_MalformedResponseIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetScriptCollection(context.Background(), "UTILITIES", "shared-scripts"); err == nil {
		t.Fatal("expected an error decoding a malformed OData response, got nil")
	}
}

func TestClient_UpdateScriptCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		want := "/api/v1/ScriptCollectionDesigntimeArtifacts(Id='shared-scripts',Version='active')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}

		var sent ScriptCollection
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if sent.Content != base64.StdEncoding.EncodeToString([]byte("new-script-bytes")) {
			t.Errorf("Content was not base64-encoded correctly: %q", sent.Content)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "shared-scripts", "Name": "Shared Scripts v2", "PackageId": "UTILITIES", "Version": "1.0.1"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	sc, err := client.UpdateScriptCollection(context.Background(), "shared-scripts", "Shared Scripts v2", []byte("new-script-bytes"))
	if err != nil {
		t.Fatalf("UpdateScriptCollection() error: %v", err)
	}
	if sc.Version != "1.0.1" {
		t.Errorf("Version = %q, want 1.0.1", sc.Version)
	}
}

func TestClient_UpdateScriptCollection_ConflictIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error": {"code": "CONFLICT", "message": {"value": "artifact is locked for editing"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	_, err := client.UpdateScriptCollection(context.Background(), "shared-scripts", "Shared Scripts v2", []byte("new-script-bytes"))

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusConflict {
		t.Fatalf("expected a 409 *apierror.Error, got %v", err)
	}
}

func TestClient_DeleteScriptCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/ScriptCollectionDesigntimeArtifacts(Id='shared-scripts',Version='active')"
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

	if err := client.DeleteScriptCollection(context.Background(), "shared-scripts"); err != nil {
		t.Fatalf("DeleteScriptCollection() error: %v", err)
	}
}

func TestClient_DeleteScriptCollection_NotFoundIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "gone"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeleteScriptCollection(context.Background(), "shared-scripts")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

func TestClient_DeployScriptCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		want := "/api/v1/DeployScriptCollectionDesigntimeArtifact?Id='shared-scripts'&Version='1.0.1'"
		if r.URL.String() != want {
			t.Errorf("url = %q, want %q", r.URL.String(), want)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeployScriptCollection(context.Background(), "shared-scripts", "1.0.1"); err != nil {
		t.Fatalf("DeployScriptCollection() error: %v", err)
	}
}

func TestClient_DeployScriptCollection_NotFoundIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "no such script collection"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeployScriptCollection(context.Background(), "missing", "active")

	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("expected a not-found *apierror.Error, got %v", err)
	}
}

package cloudintegration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetRuntimeArtifact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "metering", "Version": "1.0.0", "Status": "STARTED"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	artifact, err := client.GetRuntimeArtifact(context.Background(), "metering")
	if err != nil {
		t.Fatalf("GetRuntimeArtifact() error: %v", err)
	}
	if artifact.Status != StatusStarted {
		t.Errorf("Status = %q, want %q", artifact.Status, StatusStarted)
	}
}

func TestClient_UndeployRuntimeArtifact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/IntegrationRuntimeArtifacts('metering')"
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

	if err := client.UndeployRuntimeArtifact(context.Background(), "metering"); err != nil {
		t.Fatalf("UndeployRuntimeArtifact() error: %v", err)
	}
}

func TestClient_UndeployRuntimeArtifact_NotFoundIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "gone"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.UndeployRuntimeArtifact(context.Background(), "metering"); err == nil {
		t.Fatal("expected an error for a 404 response")
	}
}

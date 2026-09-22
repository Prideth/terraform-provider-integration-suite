package cloudintegration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_UpdatePackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/IntegrationPackages('UTILITIES')" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.UpdatePackage(context.Background(), "UTILITIES", Package{ID: "UTILITIES", Description: "updated"}); err != nil {
		t.Fatalf("UpdatePackage() error: %v", err)
	}
}

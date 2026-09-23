package apimanagementclassic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_BuildsCorrectBaseURL(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"name": "p1", "destType": "INTERNET", "useSSL": true, "trustAll": false}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL+"/")
	if _, err := client.GetAPIProvider(context.Background(), "p1"); err != nil {
		t.Fatalf("GetAPIProvider() error: %v", err)
	}

	want := "/apiportal/api/1.0/Management.svc/APIProviders('p1')"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

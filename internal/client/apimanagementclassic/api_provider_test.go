package apimanagementclassic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_CreateAPIProvider(t *testing.T) {
	var postBody []byte
	getCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/apiportal/api/1.0/Management.svc/APIProviders":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading POST body: %v", err)
			}
			postBody = body
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"name": "ES5_1", "title": "ES5", "destType": "INTERNET", "host": "sapes5.example.com", "port": 443, "useSSL": true, "trustAll": true}}`))
		case r.Method == http.MethodGet:
			getCount++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"name": "ES5_1", "destType": "INTERNET"}}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	created, err := client.CreateAPIProvider(context.Background(), APIProvider{
		Name:     "ES5_1",
		Title:    "ES5",
		DestType: "INTERNET",
		Host:     "sapes5.example.com",
		Port:     443,
		UseSSL:   true,
		TrustAll: true,
	})
	if err != nil {
		t.Fatalf("CreateAPIProvider() error: %v", err)
	}
	if created.Name != "ES5_1" {
		t.Errorf("Name = %q, want ES5_1", created.Name)
	}
	if getCount == 0 {
		t.Error("expected CreateAPIProvider to poll GET at least once to confirm visibility")
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(postBody, &decoded); err != nil {
		t.Fatalf("decoding POST body: %v", err)
	}
	if decoded["destType"] != "INTERNET" {
		t.Errorf("destType = %v, want INTERNET", decoded["destType"])
	}
}

func TestClient_CreateAPIProvider_NeverBecomesVisible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"name": "ES5_1", "destType": "INTERNET"}}`))
		case http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"lang": "en", "value": "not found"}}}`))
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.CreateAPIProvider(ctx, APIProvider{Name: "ES5_1", DestType: "INTERNET"})
	if err == nil {
		t.Fatal("expected an error when the provider never becomes visible before the context deadline")
	}
}

func TestClient_GetAPIProvider_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"lang": "en", "value": "not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	if _, err := client.GetAPIProvider(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("expected an error for a missing API provider")
	}
}

func TestClient_DeleteAPIProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		want := "/apiportal/api/1.0/Management.svc/APIProviders('ES5_1')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	if err := client.DeleteAPIProvider(context.Background(), "ES5_1"); err != nil {
		t.Fatalf("DeleteAPIProvider() error: %v", err)
	}
}

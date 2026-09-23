package securitycontent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetKeystoreEntry(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d":{"Hexalias":"736D74702E6D61696C2E7961686F6F2E636F6D","Alias":"smtp.mail.yahoo.com","KeyType":"RSA","KeySize":2048,"ValidNotBefore":"2019-10-04T00:00:00Z","ValidNotAfter":"2020-04-01T12:00:00Z"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	entry, err := client.GetKeystoreEntry(context.Background(), "smtp.mail.yahoo.com")
	if err != nil {
		t.Fatalf("GetKeystoreEntry() error: %v", err)
	}

	wantPath := "/api/v1/KeystoreEntries('736d74702e6d61696c2e7961686f6f2e636f6d')"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if entry.Alias != "smtp.mail.yahoo.com" {
		t.Errorf("Alias = %q, want smtp.mail.yahoo.com", entry.Alias)
	}
	if entry.KeyType != "RSA" {
		t.Errorf("KeyType = %q, want RSA", entry.KeyType)
	}
	if entry.KeySize != 2048 {
		t.Errorf("KeySize = %d, want 2048", entry.KeySize)
	}
}

func TestClient_GetKeystoreEntry_EscapesUnicodeAlias(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d":{"Hexalias":"c3bc6d6c617574","Alias":"ümlaut","KeyType":"RSA","KeySize":2048}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetKeystoreEntry(context.Background(), "ümlaut"); err != nil {
		t.Fatalf("GetKeystoreEntry() error: %v", err)
	}

	want := "/api/v1/KeystoreEntries('c3bc6d6c617574')"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestClient_GetKeystoreEntry_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "Not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetKeystoreEntry(context.Background(), "missing"); err == nil {
		t.Fatal("GetKeystoreEntry() error = nil, want an error for a 404 response")
	}
}

func TestClient_ListKeystoreEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/KeystoreEntries" {
			t.Errorf("path = %q, want /api/v1/KeystoreEntries", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d":{"results":[{"Hexalias":"6161","Alias":"aa","KeyType":"RSA","KeySize":2048},{"Hexalias":"6262","Alias":"bb","KeyType":"EC","KeySize":256}]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	entries, err := client.ListKeystoreEntries(context.Background())
	if err != nil {
		t.Fatalf("ListKeystoreEntries() error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Alias != "aa" || entries[1].Alias != "bb" {
		t.Errorf("entries = %+v", entries)
	}
}

func TestClient_ListKeystoreEntries_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "403", "message": {"value": "Forbidden"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.ListKeystoreEntries(context.Background()); err == nil {
		t.Fatal("ListKeystoreEntries() error = nil, want an error for a 403 response")
	}
}

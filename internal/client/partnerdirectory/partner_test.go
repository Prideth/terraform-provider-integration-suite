package partnerdirectory

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetPartner(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/Partners('PartnerZ')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Pid": "PartnerZ"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	p, err := client.GetPartner(context.Background(), "PartnerZ")
	if err != nil {
		t.Fatalf("GetPartner() error: %v", err)
	}
	if p.Pid != "PartnerZ" {
		t.Errorf("Pid = %q, want PartnerZ", p.Pid)
	}
}

func TestClient_GetPartner_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetPartner(context.Background(), "missing"); err == nil {
		t.Fatal("expected an error for a missing partner")
	}
}

func TestClient_ListPartners_FollowsPagination(t *testing.T) {
	var requests int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
		if requests == 1 {
			next := fmt.Sprintf("http://%s/api/v1/Partners?$skiptoken=page2", r.Host)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"d": {"results": [{"Pid":"A"}], "__next": %q}}`, next)))
			return
		}
		_, _ = w.Write([]byte(`{"d": {"results": [{"Pid":"B"}]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	partners, err := client.ListPartners(context.Background())
	if err != nil {
		t.Fatalf("ListPartners() error: %v", err)
	}
	if len(partners) != 2 {
		t.Fatalf("got %d partners, want 2", len(partners))
	}
	if requests != 2 {
		t.Errorf("expected 2 requests across pages, got %d", requests)
	}
}

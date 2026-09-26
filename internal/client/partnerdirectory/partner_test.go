package partnerdirectory

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
)

// SAP refuses a read of Partners by key (400, tenant test September 2026),
// so GetPartner filters the collection by Pid.
func TestClient_GetPartner(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/Partners" || r.URL.Query().Get("$filter") != "Pid eq 'PartnerZ'" {
			t.Errorf("request = %s?%s, want a $filter on Pid", r.URL.Path, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": [{"Pid": "PartnerZ"}]}}`))
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
			_, _ = fmt.Fprintf(w, `{"d": {"results": [{"Pid":"A"}], "__next": %q}}`, next)
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

// An empty filter result is a missing partner, reported as a 404.
func TestClient_GetPartner_EmptyResultIsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	_, err := New(http.DefaultClient, server.URL).GetPartner(context.Background(), "Missing")
	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Fatalf("error = %v, want a 404 apierror.Error", err)
	}
}

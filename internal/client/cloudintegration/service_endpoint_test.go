package cloudintegration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	sapthttp "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/http"
)

func TestClient_ListServiceEndpoints_ExpandsBothNestedCollections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ServiceEndpoints" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query().Get("$expand")
		if q != "EntryPoints,ApiDefinitions" {
			t.Errorf("$expand = %q, want \"EntryPoints,ApiDefinitions\"", q)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": [
			{
				"Name": "Getting Started Flow",
				"Protocol": "REST",
				"EntryPoints": {"results": [{"Name": "default", "Url": "https://tenant.example/http/getting-started", "Type": "PROD"}]},
				"ApiDefinitions": {"results": [{"Url": "https://tenant.example/api/definition.json", "Name": "oas-json"}]}
			}
		]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	endpoints, err := client.ListServiceEndpoints(context.Background(), "", "")
	if err != nil {
		t.Fatalf("ListServiceEndpoints() error: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("len(endpoints) = %d, want 1", len(endpoints))
	}

	ep := endpoints[0]
	if ep.Name != "Getting Started Flow" || ep.Protocol != "REST" {
		t.Errorf("endpoint = %+v, want Name=Getting Started Flow Protocol=REST", ep)
	}
	if len(ep.EntryPoints.Results) != 1 || ep.EntryPoints.Results[0].URL != "https://tenant.example/http/getting-started" {
		t.Errorf("EntryPoints = %+v", ep.EntryPoints)
	}
	if ep.EntryPoints.Results[0].Type != "PROD" {
		t.Errorf("EntryPoints[0].Type = %q, want PROD", ep.EntryPoints.Results[0].Type)
	}
	if len(ep.APIDefinitions.Results) != 1 || ep.APIDefinitions.Results[0].Name != "oas-json" {
		t.Errorf("APIDefinitions = %+v", ep.APIDefinitions)
	}
}

func TestClient_ListServiceEndpoints_FilterByNameAndProtocol(t *testing.T) {
	var gotFilter string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFilter = r.URL.Query().Get("$filter")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.ListServiceEndpoints(context.Background(), "Customer API", ""); err != nil {
		t.Fatalf("ListServiceEndpoints() error: %v", err)
	}
	if gotFilter != "Name eq 'Customer API'" {
		t.Errorf("$filter = %q, want \"Name eq 'Customer API'\"", gotFilter)
	}

	if _, err := client.ListServiceEndpoints(context.Background(), "", "SOAP"); err != nil {
		t.Fatalf("ListServiceEndpoints() error: %v", err)
	}
	if gotFilter != "Protocol eq 'SOAP'" {
		t.Errorf("$filter = %q, want \"Protocol eq 'SOAP'\"", gotFilter)
	}

	if _, err := client.ListServiceEndpoints(context.Background(), "Customer API", "SOAP"); err != nil {
		t.Fatalf("ListServiceEndpoints() error: %v", err)
	}
	if gotFilter != "Name eq 'Customer API' and Protocol eq 'SOAP'" {
		t.Errorf("$filter = %q, want combined filter", gotFilter)
	}
}

// TestClient_ListServiceEndpoints_FilterEscaping proves flow names
// containing apostrophes, spaces, and Unicode are safely encoded into the
// $filter expression, following OData V2's single-quote-doubling rule.
func TestClient_ListServiceEndpoints_FilterEscaping(t *testing.T) {
	var gotRawQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	name := "Customer's Ünïcödé API"
	if _, err := client.ListServiceEndpoints(context.Background(), name, ""); err != nil {
		t.Fatalf("ListServiceEndpoints() error: %v", err)
	}

	values, err := url.ParseQuery(gotRawQuery)
	if err != nil {
		t.Fatalf("parsing raw query %q: %v", gotRawQuery, err)
	}
	want := "Name eq 'Customer''s Ünïcödé API'"
	if got := values.Get("$filter"); got != want {
		t.Errorf("$filter = %q, want %q", got, want)
	}
}

func TestClient_ListServiceEndpoints_Pagination(t *testing.T) {
	var page2URL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.URL.RawQuery == "$skip=1" {
			_, _ = w.Write([]byte(`{"d": {"results": [{"Name": "Flow B", "Protocol": "SOAP", "EntryPoints": {"results": []}, "ApiDefinitions": {"results": []}}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"d": {"results": [{"Name": "Flow A", "Protocol": "REST", "EntryPoints": {"results": []}, "ApiDefinitions": {"results": []}}], "__next": "` + page2URL + `"}}`))
	}))
	defer server.Close()
	page2URL = server.URL + "/api/v1/ServiceEndpoints?$skip=1"

	client := New(http.DefaultClient, server.URL)

	endpoints, err := client.ListServiceEndpoints(context.Background(), "", "")
	if err != nil {
		t.Fatalf("ListServiceEndpoints() error: %v", err)
	}
	if len(endpoints) != 2 {
		t.Fatalf("len(endpoints) = %d, want 2 (both pages)", len(endpoints))
	}
	names := map[string]bool{endpoints[0].Name: true, endpoints[1].Name: true}
	if !names["Flow A"] || !names["Flow B"] {
		t.Errorf("endpoints = %+v, want both Flow A and Flow B across pages", endpoints)
	}
}

func TestClient_ListServiceEndpoints_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	endpoints, err := client.ListServiceEndpoints(context.Background(), "", "")
	if err != nil {
		t.Fatalf("ListServiceEndpoints() error: %v", err)
	}
	if len(endpoints) != 0 {
		t.Errorf("len(endpoints) = %d, want 0", len(endpoints))
	}
}

func TestClient_ListServiceEndpoints_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.ListServiceEndpoints(context.Background(), "", ""); err == nil {
		t.Fatal("ListServiceEndpoints() error = nil, want a decoding error for a malformed response")
	}
}

func TestClient_ListServiceEndpoints_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "403", "message": {"value": "Forbidden"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.ListServiceEndpoints(context.Background(), "", ""); err == nil {
		t.Fatal("ListServiceEndpoints() error = nil, want an error for a 403 response")
	}
}

// TestClient_ListServiceEndpoints_RetriesTransientErrors proves that a
// transient server error is retried and eventually succeeds when this
// client is used through the shared retrying sapthttp.Client, the same
// transport every resource/data source in this provider is actually
// configured with in production — not a bare http.DefaultClient like the
// other tests in this file use to isolate request-shape assertions.
func TestClient_ListServiceEndpoints_RetriesTransientErrors(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	retrying := sapthttp.New(sapthttp.Config{
		Transport:  server.Client(),
		UserAgent:  "test",
		MaxRetries: 4,
		BaseDelay:  time.Millisecond,
		MaxDelay:   5 * time.Millisecond,
	})

	client := New(retrying, server.URL)

	if _, err := client.ListServiceEndpoints(context.Background(), "", ""); err != nil {
		t.Fatalf("ListServiceEndpoints() error: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("server received %d attempts, want 3 (2 transient failures + 1 success)", got)
	}
}

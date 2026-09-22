package v2

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testStringParam struct {
	Id    string
	Value string
}

func TestGetAllPages_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": [{"Id":"A","Value":"1"},{"Id":"B","Value":"2"}]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	got, err := GetAllPages[testStringParam](context.Background(), client, "StringParameters")
	if err != nil {
		t.Fatalf("GetAllPages() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
	if got[0].Id != "A" || got[1].Id != "B" {
		t.Errorf("got %+v, want [A B]", got)
	}
}

func TestGetAllPages_FollowsNextLinks(t *testing.T) {
	var requestedPaths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.RequestURI())
		w.WriteHeader(http.StatusOK)

		switch len(requestedPaths) {
		case 1:
			next := "http://" + r.Host + "/StringParameters?$skiptoken=page2"
			_, _ = w.Write([]byte(fmt.Sprintf(`{"d": {"results": [{"Id":"A","Value":"1"}], "__next": %q}}`, next)))
		case 2:
			next := "http://" + r.Host + "/StringParameters?$skiptoken=page3"
			_, _ = w.Write([]byte(fmt.Sprintf(`{"d": {"results": [{"Id":"B","Value":"2"}], "__next": %q}}`, next)))
		default:
			_, _ = w.Write([]byte(`{"d": {"results": [{"Id":"C","Value":"3"}]}}`))
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	got, err := GetAllPages[testStringParam](context.Background(), client, "StringParameters")
	if err != nil {
		t.Fatalf("GetAllPages() error: %v", err)
	}
	if len(requestedPaths) != 3 {
		t.Fatalf("expected 3 requests (one per page), got %d: %v", len(requestedPaths), requestedPaths)
	}
	if len(got) != 3 {
		t.Fatalf("got %d entries across all pages, want 3", len(got))
	}
	ids := []string{got[0].Id, got[1].Id, got[2].Id}
	want := []string{"A", "B", "C"}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("entry %d = %q, want %q", i, ids[i], want[i])
		}
	}
}

func TestGetAllPages_EmptyCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	got, err := GetAllPages[testStringParam](context.Background(), client, "StringParameters")
	if err != nil {
		t.Fatalf("GetAllPages() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d entries, want 0", len(got))
	}
}

func TestClient_Do_FollowsAbsoluteNextURLVerbatim(t *testing.T) {
	// Regression guard for the __next-link handling in do(): an absolute
	// URL must never be rejoined with baseURL (which would produce a
	// malformed, doubled path).
	var gotPaths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL+"/api/v1")

	if _, err := client.Get(context.Background(), server.URL+"/api/v1/StringParameters?$skiptoken=abc"); err != nil {
		t.Fatalf("Get() with an absolute URL error: %v", err)
	}
	if len(gotPaths) != 1 || gotPaths[0] != "/api/v1/StringParameters" {
		t.Fatalf("got path %v, want exactly one request to /api/v1/StringParameters", gotPaths)
	}
}

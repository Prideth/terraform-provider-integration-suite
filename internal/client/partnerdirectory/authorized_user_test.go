package partnerdirectory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetCreateUpdateDeleteAuthorizedUser(t *testing.T) {
	var lastMethod, lastPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path

		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"User": "commuser1", "Pid": "PartnerZ"}}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"User": "commuser1", "Pid": "PartnerZ"}}`))
		case http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	au, err := client.GetAuthorizedUser(context.Background(), "commuser1")
	if err != nil {
		t.Fatalf("GetAuthorizedUser() error: %v", err)
	}
	if au.Pid != "PartnerZ" {
		t.Errorf("Pid = %q, want PartnerZ", au.Pid)
	}
	wantPath := "/api/v1/AuthorizedUsers('commuser1')"
	if lastPath != wantPath {
		t.Errorf("GET path = %q, want %q", lastPath, wantPath)
	}

	if _, err := client.CreateAuthorizedUser(context.Background(), AuthorizedUser{User: "commuser1", Pid: "PartnerZ"}); err != nil {
		t.Fatalf("CreateAuthorizedUser() error: %v", err)
	}
	if lastPath != "/api/v1/AuthorizedUsers" {
		t.Errorf("POST path = %q, want /api/v1/AuthorizedUsers", lastPath)
	}

	if err := client.UpdateAuthorizedUser(context.Background(), "commuser1", "PartnerY"); err != nil {
		t.Fatalf("UpdateAuthorizedUser() error: %v", err)
	}
	if lastMethod != http.MethodPut || lastPath != wantPath {
		t.Errorf("PUT: method = %q path = %q", lastMethod, lastPath)
	}

	if err := client.DeleteAuthorizedUser(context.Background(), "commuser1"); err != nil {
		t.Fatalf("DeleteAuthorizedUser() error: %v", err)
	}
	if lastMethod != http.MethodDelete || lastPath != wantPath {
		t.Errorf("DELETE: method = %q path = %q", lastMethod, lastPath)
	}
}

func TestClient_ListAuthorizedUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": [{"User":"u1","Pid":"PartnerZ"},{"User":"u2","Pid":"PartnerZ"}]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	users, err := client.ListAuthorizedUsers(context.Background(), "PartnerZ")
	if err != nil {
		t.Fatalf("ListAuthorizedUsers() error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("got %d users, want 2", len(users))
	}
}

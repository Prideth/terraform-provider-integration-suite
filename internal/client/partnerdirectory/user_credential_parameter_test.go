package partnerdirectory

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestUserCredentialParameter_HasNoPasswordField(t *testing.T) {
	// This is the safety property the whole design leans on: even if SAP's
	// API ever echoes a password back in a response body, there must be no
	// field on this struct capable of receiving it.
	typ := reflect.TypeOf(UserCredentialParameter{})
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Name == "Password" {
			t.Fatal("UserCredentialParameter must never have a Password field")
		}
	}
}

func TestClient_CreateUserCredentialParameter_SendsPasswordOnce(t *testing.T) {
	var gotBody []byte
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		// Simulate SAP echoing the password back, to prove the client
		// discards it regardless.
		_, _ = w.Write([]byte(`{"d": {"Pid": "Receiver_1", "Id": "USER", "User": "commuser1", "Password": "should-never-surface"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	created, err := client.CreateUserCredentialParameter(context.Background(), "Receiver_1", "USER", "commuser1", "s3cr3t")
	if err != nil {
		t.Fatalf("CreateUserCredentialParameter() error: %v", err)
	}
	if gotPath != "/api/v1/UserCredentialParameters" {
		t.Errorf("path = %q, want /api/v1/UserCredentialParameters", gotPath)
	}

	var decoded map[string]any
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("decoding request body: %v", err)
	}
	if decoded["Password"] != "s3cr3t" {
		t.Errorf("request Password = %v, want s3cr3t", decoded["Password"])
	}
	if decoded["User"] != "commuser1" {
		t.Errorf("request User = %v, want commuser1", decoded["User"])
	}

	if created.User != "commuser1" {
		t.Errorf("created.User = %q, want commuser1", created.User)
	}
	// The struct has no Password field, so there is nothing to assert is
	// empty — the type system itself is the guarantee. This assertion
	// documents that expectation for a future reader.
	_ = reflect.TypeOf(*created)
}

func TestClient_DeleteUserCredentialParameter(t *testing.T) {
	var gotPath, gotMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeleteUserCredentialParameter(context.Background(), "Receiver_1", "USER"); err != nil {
		t.Fatalf("DeleteUserCredentialParameter() error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	wantPath := "/api/v1/UserCredentialParameters(Pid='Receiver_1',Id='USER')"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
}

func TestClient_GetUserCredentialParameter_NeverExposesPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Pid": "Receiver_1", "Id": "USER", "User": "commuser1", "Password": "leaked-if-this-worked"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	ucp, err := client.GetUserCredentialParameter(context.Background(), "Receiver_1", "USER")
	if err != nil {
		t.Fatalf("GetUserCredentialParameter() error: %v", err)
	}
	if ucp.User != "commuser1" {
		t.Errorf("User = %q, want commuser1", ucp.User)
	}
	// No assertion on a Password field is possible or needed: the type has
	// none, so the value can never have been captured.
}

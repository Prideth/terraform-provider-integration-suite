package securitycontent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// syntheticTestPassword is an unmistakably synthetic secret used only
// against an in-process httptest server; it is never a real credential.
const syntheticTestPassword = "tf-acc-synthetic-password-not-real" // #nosec G101 -- synthetic test fixture, not a real credential

func TestClient_CreateUserCredential(t *testing.T) {
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/UserCredentials" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading POST body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Name": "BACKEND_BASIC", "User": "integration-user", "Description": "desc"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	created, err := client.CreateUserCredential(context.Background(), UserCredential{
		Name:        "BACKEND_BASIC",
		User:        "integration-user",
		Description: "desc",
	}, syntheticTestPassword)
	if err != nil {
		t.Fatalf("CreateUserCredential() error: %v", err)
	}
	if created.Name != "BACKEND_BASIC" {
		t.Errorf("Name = %q, want BACKEND_BASIC", created.Name)
	}

	var decoded map[string]any
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("decoding POST body: %v", err)
	}
	if decoded["Password"] != syntheticTestPassword {
		t.Errorf("POST body Password = %v, want the synthetic test password", decoded["Password"])
	}
}

func TestClient_CreateUserCredential_ResponseNeverExposesPassword(t *testing.T) {
	// Even if a future SAP response body echoed a password back (which SAP
	// does not document doing), UserCredential has no field to decode it
	// into. This test pins that structural guarantee down.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Name": "BACKEND_BASIC", "User": "integration-user", "Password": "` + syntheticTestPassword + `"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	created, err := client.CreateUserCredential(context.Background(), UserCredential{Name: "BACKEND_BASIC", User: "integration-user"}, syntheticTestPassword)
	if err != nil {
		t.Fatalf("CreateUserCredential() error: %v", err)
	}

	body, err := json.Marshal(created)
	if err != nil {
		t.Fatalf("marshaling created UserCredential: %v", err)
	}
	if strings.Contains(string(body), syntheticTestPassword) {
		t.Errorf("CreateUserCredential() result serializes a password: %s", body)
	}
}

func TestClient_GetUpdateDeleteUserCredential(t *testing.T) {
	var lastMethod string
	var putBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		want := "/api/v1/UserCredentials('BACKEND_BASIC')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"Name": "BACKEND_BASIC", "User": "integration-user", "Description": "desc", "Kind": "", "CompanyId": ""}}`))
		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading PUT body: %v", err)
			}
			putBody = body
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	cred, err := client.GetUserCredential(context.Background(), "BACKEND_BASIC")
	if err != nil {
		t.Fatalf("GetUserCredential() error: %v", err)
	}
	if cred.User != "integration-user" {
		t.Errorf("User = %q, want integration-user", cred.User)
	}

	err = client.UpdateUserCredential(context.Background(), UserCredential{
		Name: "BACKEND_BASIC", User: "integration-user", Description: "new",
	}, syntheticTestPassword)
	if err != nil {
		t.Fatalf("UpdateUserCredential() error: %v", err)
	}
	if lastMethod != http.MethodPut {
		t.Errorf("last method = %q, want PUT", lastMethod)
	}

	var decoded map[string]any
	if err := json.Unmarshal(putBody, &decoded); err != nil {
		t.Fatalf("decoding PUT body: %v", err)
	}
	if decoded["Password"] != syntheticTestPassword {
		t.Errorf("PUT body Password = %v, want the synthetic test password", decoded["Password"])
	}
	if decoded["Description"] != "new" {
		t.Errorf("PUT body Description = %v, want \"new\"", decoded["Description"])
	}

	if err := client.DeleteUserCredential(context.Background(), "BACKEND_BASIC"); err != nil {
		t.Fatalf("DeleteUserCredential() error: %v", err)
	}
}

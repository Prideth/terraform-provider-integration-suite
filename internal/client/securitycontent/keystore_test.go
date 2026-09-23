package securitycontent

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestEscapeKeystoreAlias_DocumentedExamples asserts the two worked
// examples from SAP's own "Trigger Mass Deletion of Keystore Entries"
// documentation byte-for-byte.
func TestEscapeKeystoreAlias_DocumentedExamples(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"semicolon", "Company A; Company B", `Company A\; Company B`},
		{"trailing backslash", `Company\A\`, `Company\\A\\`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := EscapeKeystoreAlias(c.input)
			if got != c.want {
				t.Errorf("EscapeKeystoreAlias(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

// TestJoinKeystoreAliases_DocumentedTwoAliasExample asserts SAP's second
// documented example in full: two aliases, one containing a trailing
// backslash, joined into a single Aliases string.
func TestJoinKeystoreAliases_DocumentedTwoAliasExample(t *testing.T) {
	got := joinKeystoreAliases([]string{`Company\A\`, "CompanyB"})
	want := `Company\\A\\;CompanyB`
	if got != want {
		t.Errorf("joinKeystoreAliases() = %q, want %q", got, want)
	}
}

func TestEscapeKeystoreAlias_OrderingMattersForBackslashBeforeSemicolon(t *testing.T) {
	// An alias containing both a backslash and a semicolon must have its
	// backslash escaped first; escaping in the other order would
	// re-escape the backslash this function itself inserts for the
	// semicolon, corrupting the result.
	got := EscapeKeystoreAlias(`a\;b`)
	want := `a\\\;b`
	if got != want {
		t.Errorf("EscapeKeystoreAlias(%q) = %q, want %q", `a\;b`, got, want)
	}
}

func TestClient_DeleteKeystoreEntries_ExactRequestBody(t *testing.T) {
	var gotPath, gotMethod, gotQuery string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.RawQuery
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeleteKeystoreEntries(context.Background(), []string{"alias1", "alias2", "alias3"})
	if err != nil {
		t.Fatalf("DeleteKeystoreEntries() error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if gotPath != "/api/v1/KeystoreResources('system')" {
		t.Errorf("path = %q, want /api/v1/KeystoreResources('system')", gotPath)
	}
	if gotQuery != "deleteEntries=true" {
		t.Errorf("query = %q, want deleteEntries=true", gotQuery)
	}

	want := `{"Aliases":"alias1;alias2;alias3"}`
	if string(gotBody) != want {
		t.Errorf("request body = %s, want %s", gotBody, want)
	}
}

// TestClient_DeleteKeystoreEntries_SingleAlias is the critical
// destructive-safety regression test: a Terraform resource destroying only
// the one alias it owns must never construct a request body that also
// references any other alias.
func TestClient_DeleteKeystoreEntries_SingleAlias(t *testing.T) {
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeleteKeystoreEntries(context.Background(), []string{"tf-acc-cert-only-this-one"})
	if err != nil {
		t.Fatalf("DeleteKeystoreEntries() error: %v", err)
	}

	want := `{"Aliases":"tf-acc-cert-only-this-one"}`
	if string(gotBody) != want {
		t.Errorf("request body = %s, want %s (must contain exactly one alias)", gotBody, want)
	}
}

func TestClient_DeleteKeystoreEntries_RequiresAtLeastOneAlias(t *testing.T) {
	client := New(http.DefaultClient, "https://example.invalid")

	if err := client.DeleteKeystoreEntries(context.Background(), nil); err == nil {
		t.Error("DeleteKeystoreEntries(nil) error = nil, want an error")
	}
	if err := client.DeleteKeystoreEntries(context.Background(), []string{}); err == nil {
		t.Error("DeleteKeystoreEntries([]string{}) error = nil, want an error")
	}
}

func TestClient_DeleteKeystoreEntries_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "403", "message": {"value": "Cannot delete SAP-owned entry"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.DeleteKeystoreEntries(context.Background(), []string{"sap_managed_entry"})
	if err == nil {
		t.Fatal("DeleteKeystoreEntries() error = nil, want an error for a 403 response (SAP-owned entry)")
	}
}

func TestClient_DeleteKeystoreEntries_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "Not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.DeleteKeystoreEntries(context.Background(), []string{"missing"}); err == nil {
		t.Fatal("DeleteKeystoreEntries() error = nil, want an error for a 404 response")
	}
}

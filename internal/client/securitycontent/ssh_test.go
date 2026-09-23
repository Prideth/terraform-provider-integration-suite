package securitycontent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetSSHPublicKey(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ssh-rsa AAAAB3NzaC1yc2E dKoanTZDcJxo5 public key for alias baltimore cybertrust root\n"))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	got, err := client.GetSSHPublicKey(context.Background(), "baltimore cybertrust root")
	if err != nil {
		t.Fatalf("GetSSHPublicKey() error: %v", err)
	}

	// SAP's own documented example.
	wantPath := "/api/v1/KeystoreEntries('62616c74696d6f7265206379626572747275737420726f6f74')/Sshkey/$value"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if len(got) == 0 {
		t.Error("GetSSHPublicKey() returned empty body")
	}
}

func TestClient_GetSSHPublicKey_UnsupportedKeyType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"code": "400", "message": {"value": "EC keys are not supported for SSH export"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetSSHPublicKey(context.Background(), "ec-key"); err == nil {
		t.Fatal("GetSSHPublicKey() error = nil, want an error for an unsupported (EC) key type")
	}
}

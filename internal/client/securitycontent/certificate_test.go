package securitycontent

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testCertificatePEM = `-----BEGIN CERTIFICATE-----
MIIITDCCBzSgAwIBAgIQDBxHETUvPf18xKE7JdDoNjANBgkqhkiG9w0BAQsFADBw
F8jCgaR+bTz+hYoLtFN9mow1kr5jNgcn68HALdKF8SluCyZS5jFbDOM4HQay16pU
J++y5pvlBMtzQ571O7itrSQ32praT6whMlCBy9pAnJIKzVM4IDM5H1drHc6shhuo
-----END CERTIFICATE-----
`

func TestClient_PutCertificate(t *testing.T) {
	var gotPath, gotMethod, gotContentType string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.PutCertificate(context.Background(), "mycertificate", []byte(testCertificatePEM))
	if err != nil {
		t.Fatalf("PutCertificate() error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}

	// SAP's own documented example.
	wantPath := "/api/v1/CertificateResources('6d796365727469666963617465')/$value"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if gotContentType != "application/pkix-cert" {
		t.Errorf("Content-Type = %q, want application/pkix-cert", gotContentType)
	}
	if string(gotBody) != testCertificatePEM {
		t.Errorf("request body sent to SAP does not match the input PEM content exactly")
	}
}

func TestClient_PutCertificate_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error": {"code": "409", "message": {"value": "Conflict"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.PutCertificate(context.Background(), "dup", []byte(testCertificatePEM))
	if err == nil {
		t.Fatal("PutCertificate() error = nil, want an error for a 409 response")
	}
}

func TestClient_PutCertificate_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": {"code": "403", "message": {"value": "Forbidden - SAP-owned entry"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.PutCertificate(context.Background(), "sap_owned", []byte(testCertificatePEM))
	if err == nil {
		t.Fatal("PutCertificate() error = nil, want an error for a 403 response")
	}
}

func TestClient_GetCertificate(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/pkix-cert")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testCertificatePEM))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	got, err := client.GetCertificate(context.Background(), "smtp.mail.yahoo.com")
	if err != nil {
		t.Fatalf("GetCertificate() error: %v", err)
	}

	wantPath := "/api/v1/KeystoreEntries('736d74702e6d61696c2e7961686f6f2e636f6d')/Certificate/$value"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if string(got) != testCertificatePEM {
		t.Errorf("GetCertificate() body does not match")
	}
}

func TestClient_GetCertificate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "Not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if _, err := client.GetCertificate(context.Background(), "missing"); err == nil {
		t.Fatal("GetCertificate() error = nil, want an error for a 404 response")
	}
}

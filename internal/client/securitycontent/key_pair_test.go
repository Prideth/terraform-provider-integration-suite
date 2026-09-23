package securitycontent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GenerateKeyPair_ExactRequestBody(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.GenerateKeyPair(context.Background(), KeyPairGenerationRequest{
		Alias:              "cust_alias",
		KeyType:            "RSA",
		SignatureAlgorithm: "SHA-256/RSA",
		KeySize:            2048,
		CommonName:         "cpi.test.wdf.sap.com",
		OrganizationUnit:   "My Organization Unit",
		Organization:       "My Organization",
		Locality:           "My Location",
		State:              "My State",
		Country:            "DE",
		Email:              "mymail@myorganization.com",
	})
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/v1/KeyPairGenerationRequests" {
		t.Errorf("path = %q, want /api/v1/KeyPairGenerationRequests", gotPath)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("decoding request body: %v", err)
	}

	if decoded["Alias"] != "cust_alias" {
		t.Errorf("Alias = %v, want cust_alias", decoded["Alias"])
	}
	if decoded["CommonName"] != "cpi.test.wdf.sap.com" {
		t.Errorf("CommonName = %v", decoded["CommonName"])
	}
	if decoded["Country"] != "DE" {
		t.Errorf("Country = %v, want DE", decoded["Country"])
	}
	if decoded["SignatureAlgorithm"] != "SHA-256/RSA" {
		t.Errorf("SignatureAlgorithm = %v, want SHA-256/RSA", decoded["SignatureAlgorithm"])
	}
	if _, ok := decoded["Hexalias"]; !ok {
		t.Error("request body does not contain the required Hexalias property")
	}
	if _, ok := decoded["ValidNotBefore"]; ok {
		t.Error("ValidNotBefore should be omitted when not specified, letting SAP's default apply")
	}
}

func TestClient_GenerateKeyPair_EncodesValidityDates(t *testing.T) {
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	before := time.UnixMilli(1469685029780).UTC()
	err := client.GenerateKeyPair(context.Background(), KeyPairGenerationRequest{
		Alias:          "cust_alias",
		CommonName:     "cpi.test.wdf.sap.com",
		Country:        "DE",
		ValidNotBefore: &before,
	})
	if err != nil {
		t.Fatalf("GenerateKeyPair() error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("decoding request body: %v", err)
	}

	// SAP's own documented example wire value for this exact millisecond value.
	want := "/Date(1469685029780)/"
	if decoded["ValidNotBefore"] != want {
		t.Errorf("ValidNotBefore = %v, want %v", decoded["ValidNotBefore"], want)
	}
}

func TestClient_GenerateKeyPair_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error": {"code": "409", "message": {"value": "Conflict"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.GenerateKeyPair(context.Background(), KeyPairGenerationRequest{
		Alias: "dup", CommonName: "cn", Country: "DE",
	})
	if err == nil {
		t.Fatal("GenerateKeyPair() error = nil, want an error for a 409 response")
	}
}

func TestClient_GenerateKeyPair_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"code": "400", "message": {"value": "Invalid KeyAlgorithmParameter"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.GenerateKeyPair(context.Background(), KeyPairGenerationRequest{
		Alias: "x", CommonName: "cn", Country: "DE", KeyType: "EC",
	})
	if err == nil {
		t.Fatal("GenerateKeyPair() error = nil, want an error for a 400 response")
	}
}

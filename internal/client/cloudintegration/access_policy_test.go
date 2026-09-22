package cloudintegration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateAccessPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/AccessPolicies" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"Id": "1", "RoleName": "UTILITIES_ARCHITECT", "Description": "desc"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	policy, err := client.CreateAccessPolicy(context.Background(), AccessPolicy{RoleName: "UTILITIES_ARCHITECT", Description: "desc"})
	if err != nil {
		t.Fatalf("CreateAccessPolicy() error: %v", err)
	}
	if policy.ID != "1" {
		t.Errorf("ID = %q, want 1", policy.ID)
	}
}

func TestClient_GetUpdateDeleteAccessPolicy(t *testing.T) {
	var lastMethod string
	var patchBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		want := "/api/v1/AccessPolicies('1')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"Id": "1", "RoleName": "UTILITIES_ARCHITECT", "ReconciliationStatus": "SUCCESS"}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading PATCH body: %v", err)
			}
			patchBody = body
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	policy, err := client.GetAccessPolicy(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetAccessPolicy() error: %v", err)
	}
	if policy.ReconciliationStatus != "SUCCESS" {
		t.Errorf("ReconciliationStatus = %q, want SUCCESS", policy.ReconciliationStatus)
	}

	if err := client.UpdateAccessPolicy(context.Background(), "1", "new"); err != nil {
		t.Fatalf("UpdateAccessPolicy() error: %v", err)
	}
	if lastMethod != http.MethodPatch {
		t.Errorf("last method = %q, want PATCH", lastMethod)
	}

	var decoded map[string]any
	if err := json.Unmarshal(patchBody, &decoded); err != nil {
		t.Fatalf("decoding PATCH body: %v", err)
	}
	if len(decoded) != 1 {
		t.Errorf("PATCH body has %d fields, want exactly 1 (Description): %s", len(decoded), patchBody)
	}
	if decoded["Description"] != "new" {
		t.Errorf("PATCH body Description = %v, want \"new\": %s", decoded["Description"], patchBody)
	}
	if _, hasRoleName := decoded["RoleName"]; hasRoleName {
		t.Errorf("PATCH body must not include the immutable RoleName field: %s", patchBody)
	}

	if err := client.DeleteAccessPolicy(context.Background(), "1"); err != nil {
		t.Fatalf("DeleteAccessPolicy() error: %v", err)
	}
}

func TestClient_GetAccessPolicyReference(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/AccessPolicies('1')/ArtifactReferences('ref-1')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "ref-1", "ArtifactType": "IntegrationFlow", "Attribute": "Name", "Operator": "EQUALS", "Value": "metering"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	ref, err := client.GetAccessPolicyReference(context.Background(), "1", "ref-1")
	if err != nil {
		t.Fatalf("GetAccessPolicyReference() error: %v", err)
	}
	if ref.Value != "metering" {
		t.Errorf("Value = %q, want metering", ref.Value)
	}
}

func TestClient_AccessPolicyReferences(t *testing.T) {
	var createPath, deletePath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createPath = r.URL.Path
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"Id": "ref-1", "ArtifactType": "IntegrationFlow", "Attribute": "Name", "Operator": "EQUALS", "Value": "metering"}}`))
		case http.MethodDelete:
			deletePath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	ref, err := client.CreateAccessPolicyReference(context.Background(), "1", AccessPolicyReference{
		ArtifactType: "IntegrationFlow",
		Attribute:    "Name",
		Operator:     "EQUALS",
		Value:        "metering",
	})
	if err != nil {
		t.Fatalf("CreateAccessPolicyReference() error: %v", err)
	}
	if ref.ID != "ref-1" {
		t.Errorf("ID = %q, want ref-1", ref.ID)
	}

	wantCreatePath := "/api/v1/AccessPolicies('1')/ArtifactReferences"
	if createPath != wantCreatePath {
		t.Errorf("create path = %q, want %q", createPath, wantCreatePath)
	}

	if err := client.DeleteAccessPolicyReference(context.Background(), "1", "ref-1"); err != nil {
		t.Fatalf("DeleteAccessPolicyReference() error: %v", err)
	}

	wantDeletePath := "/api/v1/AccessPolicies('1')/ArtifactReferences('ref-1')"
	if deletePath != wantDeletePath {
		t.Errorf("delete path = %q, want %q", deletePath, wantDeletePath)
	}
}

func TestSupportedArtifactTypes_ContainsDocumentedTypes(t *testing.T) {
	want := map[string]bool{
		"IntegrationFlow": true, "ODataAPI": true, "RestAPI": true, "SoapAPI": true,
		"ScriptCollection": true, "ValueMapping": true, "MessageMapping": true,
		"MessageQueue": true, "GlobalDataStore": true, "GlobalVariable": true,
	}
	if len(SupportedArtifactTypes) != len(want) {
		t.Fatalf("SupportedArtifactTypes has %d entries, want %d", len(SupportedArtifactTypes), len(want))
	}
	for _, v := range SupportedArtifactTypes {
		if !want[v] {
			t.Errorf("unexpected artifact type %q", v)
		}
	}
}

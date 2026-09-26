package apimanagementclassic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateKeyValueMap(t *testing.T) {
	var postBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/apiportal/api/1.0/Management.svc/GenericKeyMapEntries" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		postBody = body
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"name": "apim.oc.instance.token", "scope": "APIPROXY", "scopeId": "SampleAPI", "isEncrypted": false, "genericKeyMapEntryValues": {"__deferred": {"uri": "x"}}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	created, err := client.CreateKeyValueMap(context.Background(), KeyValueMap{
		Name:    "apim.oc.instance.token",
		Scope:   "APIPROXY",
		ScopeID: "SampleAPI",
		Entries: []KeyValueMapEntry{{Name: "default", Value: "secret"}},
	})
	if err != nil {
		t.Fatalf("CreateKeyValueMap() error: %v", err)
	}
	if len(created.Entries) != 1 || created.Entries[0].Value != "secret" {
		t.Errorf("Entries = %v, want the sent entry (SAP answers with only a __deferred link)", created.Entries)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(postBody, &decoded); err != nil {
		t.Fatalf("decoding POST body: %v", err)
	}
	if decoded["isEncrypted"] != false {
		t.Errorf("isEncrypted = %v, want false", decoded["isEncrypted"])
	}
}

func TestClient_GetDeleteKeyValueMap(t *testing.T) {
	var lastMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		mapPath := "/apiportal/api/1.0/Management.svc/GenericKeyMapEntries(name='kvm1',scope='APIPROXY',scopeId='SampleAPI')"
		switch {
		case r.Method == http.MethodGet && r.URL.Path == mapPath:
			// As a tenant returned it in September 2026: entries only as a link.
			_, _ = w.Write([]byte(`{"d": {"name": "kvm1", "scope": "APIPROXY", "scopeId": "SampleAPI", "isEncrypted": false, "genericKeyMapEntryValues": {"__deferred": {"uri": "x"}}}}`))
		case r.Method == http.MethodGet && r.URL.Path == mapPath+"/genericKeyMapEntryValues":
			_, _ = w.Write([]byte(`{"d": {"results": [{"name": "default", "mapName": "kvm1", "value": "v1", "scope": "APIPROXY", "scopeId": "SampleAPI"}]}}`))
		case r.Method == http.MethodDelete && r.URL.Path == mapPath:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	kvm, err := client.GetKeyValueMap(context.Background(), "kvm1", "APIPROXY", "SampleAPI")
	if err != nil {
		t.Fatalf("GetKeyValueMap() error: %v", err)
	}
	if kvm.Name != "kvm1" {
		t.Errorf("Name = %q, want kvm1", kvm.Name)
	}
	if len(kvm.Entries) != 1 || kvm.Entries[0].Name != "default" || kvm.Entries[0].Value != "v1" {
		t.Errorf("Entries = %v, want default=v1 from genericKeyMapEntryValues", kvm.Entries)
	}

	if err := client.DeleteKeyValueMap(context.Background(), "kvm1", "APIPROXY", "SampleAPI"); err != nil {
		t.Fatalf("DeleteKeyValueMap() error: %v", err)
	}
	if lastMethod != http.MethodDelete {
		t.Errorf("lastMethod = %s, want DELETE", lastMethod)
	}
}

func TestClient_GetKeyValueMap_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"lang": "en", "value": "not found"}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	if _, err := client.GetKeyValueMap(context.Background(), "missing", "APIPROXY", "SampleAPI"); err == nil {
		t.Fatal("expected an error for a missing key value map")
	}
}

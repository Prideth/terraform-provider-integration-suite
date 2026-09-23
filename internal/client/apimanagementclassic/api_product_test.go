package apimanagementclassic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateAPIProduct(t *testing.T) {
	var postBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/apiportal/api/1.0/Management.svc/APIProducts" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		postBody = body
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"name": "SampleProduct", "version": "1", "title": "SampleProduct", "scope": "", "status_code": "PUBLISHED", "isPublished": false, "isRestricted": false, "apiProxies": [{"__metadata": {"uri": "APIProxies(name='SampleAPI')"}}]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	created, err := client.CreateAPIProduct(context.Background(), APIProduct{
		Name:          "SampleProduct",
		Version:       "1",
		Title:         "SampleProduct",
		StatusCode:    "PUBLISHED",
		ApiProxyNames: []string{"SampleAPI"},
	})
	if err != nil {
		t.Fatalf("CreateAPIProduct() error: %v", err)
	}
	if len(created.ApiProxyNames) != 1 || created.ApiProxyNames[0] != "SampleAPI" {
		t.Errorf("ApiProxyNames = %v, want [SampleAPI]", created.ApiProxyNames)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(postBody, &decoded); err != nil {
		t.Fatalf("decoding POST body: %v", err)
	}
	proxies, ok := decoded["apiProxies"].([]interface{})
	if !ok || len(proxies) != 1 {
		t.Fatalf("apiProxies in request body = %v", decoded["apiProxies"])
	}
}

func TestClient_GetUpdateDeleteAPIProduct(t *testing.T) {
	var lastMethod string
	var putBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		want := "/apiportal/api/1.0/Management.svc/APIProducts('SampleProduct')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"name": "SampleProduct", "title": "SampleProduct", "status_code": "PUBLISHED"}}`))
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			putBody = body
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	product, err := client.GetAPIProduct(context.Background(), "SampleProduct")
	if err != nil {
		t.Fatalf("GetAPIProduct() error: %v", err)
	}
	if product.Title != "SampleProduct" {
		t.Errorf("Title = %q, want SampleProduct", product.Title)
	}

	if err := client.UpdateAPIProduct(context.Background(), APIProduct{
		Name:        "SampleProduct",
		Title:       "SampleProduct",
		StatusCode:  "PUBLISHED",
		IsPublished: true,
	}); err != nil {
		t.Fatalf("UpdateAPIProduct() error: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(putBody, &decoded); err != nil {
		t.Fatalf("decoding PUT body: %v", err)
	}
	if _, present := decoded["apiProxies"]; present {
		t.Error("Update payload should never include apiProxies (no confirmed update path for the association)")
	}

	if err := client.DeleteAPIProduct(context.Background(), "SampleProduct"); err != nil {
		t.Fatalf("DeleteAPIProduct() error: %v", err)
	}
	if lastMethod != http.MethodDelete {
		t.Errorf("lastMethod = %s, want DELETE", lastMethod)
	}
}

func TestClient_APIProductAdditionalProperty(t *testing.T) {
	var lastMethod, lastPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"entityId": "SampleProduct", "name": "key1", "value": "val1"}}`))
		case http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	if err := client.CreateAPIProductAdditionalProperty(context.Background(), APIProductAdditionalProperty{
		EntityID: "SampleProduct", Name: "key1", Value: "val1",
	}); err != nil {
		t.Fatalf("CreateAPIProductAdditionalProperty() error: %v", err)
	}

	if err := client.UpdateAPIProductAdditionalProperty(context.Background(), "SampleProduct", "key1", "val2"); err != nil {
		t.Fatalf("UpdateAPIProductAdditionalProperty() error: %v", err)
	}
	wantPath := "/apiportal/api/1.0/Management.svc/APIProductAdditionalProperties(entityId='SampleProduct',name='key1')"
	if lastPath != wantPath {
		t.Errorf("path = %q, want %q", lastPath, wantPath)
	}
	if lastMethod != http.MethodPut {
		t.Errorf("lastMethod = %s, want PUT", lastMethod)
	}

	if err := client.DeleteAPIProductAdditionalProperty(context.Background(), "SampleProduct", "key1"); err != nil {
		t.Fatalf("DeleteAPIProductAdditionalProperty() error: %v", err)
	}
}

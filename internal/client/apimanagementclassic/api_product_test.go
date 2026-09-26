package apimanagementclassic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// deferredProduct is the product as a tenant returned it in September 2026:
// navigation properties are __deferred objects, not arrays.
const deferredProduct = `{"d": {"__metadata": {"type": "apiportal.APIProduct"}, "name": "SampleProduct", "version": "1", "title": "SampleProduct", "description": null, "scope": "", "status_code": "PUBLISHED", "isPublished": true, "isRestricted": false, "quotaCount": null, "quotaInterval": null, "quotaTimeUnit": null, "life_cycle": {"created_at": "/Date(1790423910729)/"}, "apiProxies": {"__deferred": {"uri": "https://host/APIProducts('SampleProduct')/apiProxies"}}, "apiResources": {"__deferred": {"uri": "https://host/APIProducts('SampleProduct')/apiResources"}}, "additionalProperties": {"__deferred": {"uri": "https://host/APIProducts('SampleProduct')/additionalProperties"}}}}`

func TestClient_CreateAPIProduct(t *testing.T) {
	var postBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/apiportal/api/1.0/Management.svc/APIProducts" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		postBody = body
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(deferredProduct))
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
		t.Errorf("ApiProxyNames = %v, want the sent [SampleAPI]", created.ApiProxyNames)
	}
	if !created.IsPublished {
		t.Error("IsPublished = false, want the value SAP returned")
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(postBody, &decoded); err != nil {
		t.Fatalf("decoding POST body: %v", err)
	}
	proxies, ok := decoded["apiProxies"].([]interface{})
	if !ok || len(proxies) != 1 {
		t.Fatalf("apiProxies in request body = %v", decoded["apiProxies"])
	}
	if decoded["status_code"] != "PUBLISHED" {
		t.Errorf("status_code in request body = %v, want PUBLISHED", decoded["status_code"])
	}
}

func TestClient_GetDeleteAPIProduct(t *testing.T) {
	var lastMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		want := "/apiportal/api/1.0/Management.svc/APIProducts('SampleProduct')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(deferredProduct))
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
	if product.Title != "SampleProduct" || product.StatusCode != "PUBLISHED" {
		t.Errorf("product = %+v, want title SampleProduct and status PUBLISHED", product)
	}

	if err := client.DeleteAPIProduct(context.Background(), "SampleProduct"); err != nil {
		t.Fatalf("DeleteAPIProduct() error: %v", err)
	}
	if lastMethod != http.MethodDelete {
		t.Errorf("lastMethod = %s, want DELETE", lastMethod)
	}
}

// SAP rejects a deep-inserted property without the owning product's name
// (400 ADDITIONAL_PROPERTY_ENTITY_ID_MATCH_ERROR), so the create body must
// carry it in every property.
func TestClient_CreateAPIProduct_PropertiesCarryEntityID(t *testing.T) {
	var postBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &postBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(deferredProduct))
	}))
	defer server.Close()

	_, err := New(http.DefaultClient, server.URL).CreateAPIProduct(context.Background(), APIProduct{
		Name:                 "SampleProduct",
		StatusCode:           "PUBLISHED",
		ApiProxyNames:        []string{"SampleAPI"},
		AdditionalProperties: []APIProductAdditionalProperty{{Name: "team", Value: "integration"}},
	})
	if err != nil {
		t.Fatalf("CreateAPIProduct() error: %v", err)
	}
	props, _ := postBody["additionalProperties"].([]interface{})
	if len(props) != 1 {
		t.Fatalf("additionalProperties sent = %v, want one", postBody["additionalProperties"])
	}
	prop := props[0].(map[string]interface{})
	if prop["entityId"] != "SampleProduct" || prop["name"] != "team" || prop["value"] != "integration" {
		t.Errorf("property sent = %v, want entityId SampleProduct, name team, value integration", prop)
	}
}

func TestClient_GetAPIProductNavigation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := "/apiportal/api/1.0/Management.svc/APIProducts('SampleProduct')"
		switch r.URL.Path {
		case base + "/apiProxies":
			// Trimmed from a tenant answer, September 2026.
			_, _ = w.Write([]byte(`{"d": {"results": [{"__metadata": {"type": "apiportal.APIProxy"}, "name": "test", "state": "DEPLOYED", "status_code": "REGISTERED", "life_cycle": {"created_at": "/Date(1790423031733)/"}}]}}`))
		case base + "/additionalProperties":
			_, _ = w.Write([]byte(`{"d": {"results": [{"entityId": "SampleProduct", "name": "team", "value": "integration"}]}}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()
	client := New(http.DefaultClient, server.URL)

	names, err := client.GetAPIProductProxyNames(context.Background(), "SampleProduct")
	if err != nil {
		t.Fatalf("GetAPIProductProxyNames() error: %v", err)
	}
	if len(names) != 1 || names[0] != "test" {
		t.Errorf("names = %v, want [test]", names)
	}

	props, err := client.GetAPIProductAdditionalProperties(context.Background(), "SampleProduct")
	if err != nil {
		t.Fatalf("GetAPIProductAdditionalProperties() error: %v", err)
	}
	if len(props) != 1 || props[0].Name != "team" || props[0].Value != "integration" {
		t.Errorf("props = %+v, want team=integration", props)
	}
}

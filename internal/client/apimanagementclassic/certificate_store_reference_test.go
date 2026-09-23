package apimanagementclassic

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateCertificateStoreReference(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"name":"reference-name","certificateStoreName":"store-name"}` {
			t.Errorf("unexpected body: %s", body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d": {"name": "reference-name", "certificateStoreName": "store-name", "storeType": "TRUSTSTORE"}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	created, err := client.CreateCertificateStoreReference(context.Background(), CertificateStoreReference{
		Name:                 "reference-name",
		CertificateStoreName: "store-name",
	})
	if err != nil {
		t.Fatalf("CreateCertificateStoreReference() error: %v", err)
	}
	if created.StoreType != "TRUSTSTORE" {
		t.Errorf("StoreType = %q, want TRUSTSTORE", created.StoreType)
	}
}

func TestClient_GetUpdateDeleteCertificateStoreReference(t *testing.T) {
	var lastMethod string
	var putBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		want := "/apiportal/api/1.0/Management.svc/CertificateStoreReferences('reference_ts_1')"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"name": "reference_ts_1", "certificateStoreName": "ts", "storeType": "TRUSTSTORE"}}`))
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

	ref, err := client.GetCertificateStoreReference(context.Background(), "reference_ts_1")
	if err != nil {
		t.Fatalf("GetCertificateStoreReference() error: %v", err)
	}
	if ref.CertificateStoreName != "ts" {
		t.Errorf("CertificateStoreName = %q, want ts", ref.CertificateStoreName)
	}

	if err := client.UpdateCertificateStoreReference(context.Background(), "reference_ts_1", "updated-store-name"); err != nil {
		t.Fatalf("UpdateCertificateStoreReference() error: %v", err)
	}
	if string(putBody) != `{"certificateStoreName":"updated-store-name"}` {
		t.Errorf("unexpected PUT body: %s", putBody)
	}

	if err := client.DeleteCertificateStoreReference(context.Background(), "reference_ts_1"); err != nil {
		t.Fatalf("DeleteCertificateStoreReference() error: %v", err)
	}
	if lastMethod != http.MethodDelete {
		t.Errorf("lastMethod = %s, want DELETE", lastMethod)
	}
}

func TestClient_GetCertificateStoreReference_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "NO_SUCH_CERTIFICATE_STORE_REFERENCE_EXISTS", "message": {"lang": "en", "value": "No such certificate store reference exists."}}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	if _, err := client.GetCertificateStoreReference(context.Background(), "missing"); err == nil {
		t.Fatal("expected an error for a missing certificate store reference")
	}
}

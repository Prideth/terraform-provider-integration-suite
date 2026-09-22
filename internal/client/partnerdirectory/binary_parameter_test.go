package partnerdirectory

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetCreateUpdateDeleteBinaryParameter(t *testing.T) {
	content := base64.StdEncoding.EncodeToString([]byte("<xsd>schema</xsd>"))
	var lastMethod, lastPath string
	var lastBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		lastBody = body

		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"Pid": "PartnerZ", "Id": "OrderSchema", "ContentType": "xsd", "Value": "` + content + `"}}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"Pid": "PartnerZ", "Id": "OrderSchema", "ContentType": "xsd", "Value": "` + content + `"}}`))
		case http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	bp, err := client.GetBinaryParameter(context.Background(), "PartnerZ", "OrderSchema")
	if err != nil {
		t.Fatalf("GetBinaryParameter() error: %v", err)
	}
	if bp.Value != content {
		t.Errorf("Value = %q, want %q", bp.Value, content)
	}
	if bp.ContentType != "xsd" {
		t.Errorf("ContentType = %q, want xsd", bp.ContentType)
	}

	created, err := client.CreateBinaryParameter(context.Background(), "PartnerZ", "OrderSchema", "xsd", []byte("<xsd>schema</xsd>"))
	if err != nil {
		t.Fatalf("CreateBinaryParameter() error: %v", err)
	}
	if created.Id != "OrderSchema" {
		t.Errorf("created.Id = %q, want OrderSchema", created.Id)
	}

	if err := client.UpdateBinaryParameter(context.Background(), "PartnerZ", "OrderSchema", "zip", []byte("<xsd>schema</xsd>")); err != nil {
		t.Fatalf("UpdateBinaryParameter() error: %v", err)
	}
	if lastMethod != http.MethodPut {
		t.Errorf("last method = %q, want PUT", lastMethod)
	}
	wantPath := "/api/v1/BinaryParameters(Pid='PartnerZ',Id='OrderSchema')"
	if lastPath != wantPath {
		t.Errorf("PUT path = %q, want %q", lastPath, wantPath)
	}
	if len(lastBody) == 0 {
		t.Error("expected a non-empty PUT body")
	}

	if err := client.DeleteBinaryParameter(context.Background(), "PartnerZ", "OrderSchema"); err != nil {
		t.Fatalf("DeleteBinaryParameter() error: %v", err)
	}
}

func TestClient_ListBinaryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": [{"Pid":"PartnerZ","Id":"A","ContentType":"xml","Value":"YQ=="}]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	params, err := client.ListBinaryParameters(context.Background(), "PartnerZ")
	if err != nil {
		t.Fatalf("ListBinaryParameters() error: %v", err)
	}
	if len(params) != 1 {
		t.Fatalf("got %d params, want 1", len(params))
	}
}

func TestMaxBinaryParameterValueBytes(t *testing.T) {
	if MaxBinaryParameterValueBytes != 260*1024 {
		t.Errorf("MaxBinaryParameterValueBytes = %d, want %d (260 KB)", MaxBinaryParameterValueBytes, 260*1024)
	}
}

func TestDocumentedBinaryParameterContentTypes_ContainsKnownValues(t *testing.T) {
	want := map[string]bool{
		"xml": true, "xsl": true, "xsd": true, "json": true, "text": true,
		"zip": true, "gz": true, "zlib": true, "crt": true,
	}
	if len(DocumentedBinaryParameterContentTypes) != len(want) {
		t.Fatalf("got %d documented content types, want %d", len(DocumentedBinaryParameterContentTypes), len(want))
	}
	for _, v := range DocumentedBinaryParameterContentTypes {
		if !want[v] {
			t.Errorf("unexpected documented content type %q", v)
		}
	}
}

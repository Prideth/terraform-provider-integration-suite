package partnerdirectory

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetCreateUpdateDeleteStringParameter(t *testing.T) {
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
			_, _ = w.Write([]byte(`{"d": {"Pid": "PartnerZ", "Id": "ReceiverAddress", "Value": "https://receiver.example.com"}}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"Pid": "PartnerZ", "Id": "ReceiverAddress", "Value": "https://receiver.example.com"}}`))
		case http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	sp, err := client.GetStringParameter(context.Background(), "PartnerZ", "ReceiverAddress")
	if err != nil {
		t.Fatalf("GetStringParameter() error: %v", err)
	}
	wantPath := "/api/v1/StringParameters(Pid='PartnerZ',Id='ReceiverAddress')"
	if lastPath != wantPath {
		t.Errorf("GET path = %q, want %q", lastPath, wantPath)
	}
	if sp.Value != "https://receiver.example.com" {
		t.Errorf("Value = %q, want the receiver URL", sp.Value)
	}

	created, err := client.CreateStringParameter(context.Background(), StringParameter{
		Pid: "PartnerZ", Id: "ReceiverAddress", Value: "https://receiver.example.com",
	})
	if err != nil {
		t.Fatalf("CreateStringParameter() error: %v", err)
	}
	if created.Pid != "PartnerZ" {
		t.Errorf("created.Pid = %q, want PartnerZ", created.Pid)
	}
	if lastPath != "/api/v1/StringParameters" {
		t.Errorf("POST path = %q, want /api/v1/StringParameters", lastPath)
	}

	if err := client.UpdateStringParameter(context.Background(), "PartnerZ", "ReceiverAddress", "https://new.example.com"); err != nil {
		t.Fatalf("UpdateStringParameter() error: %v", err)
	}
	if lastMethod != http.MethodPut {
		t.Errorf("last method = %q, want PUT", lastMethod)
	}
	if lastPath != wantPath {
		t.Errorf("PUT path = %q, want %q", lastPath, wantPath)
	}
	var decoded map[string]any
	if err := json.Unmarshal(lastBody, &decoded); err != nil {
		t.Fatalf("decoding PUT body: %v", err)
	}
	if len(decoded) != 1 || decoded["Value"] != "https://new.example.com" {
		t.Errorf("PUT body = %s, want exactly {\"Value\":\"https://new.example.com\"}", lastBody)
	}

	if err := client.DeleteStringParameter(context.Background(), "PartnerZ", "ReceiverAddress"); err != nil {
		t.Fatalf("DeleteStringParameter() error: %v", err)
	}
	if lastMethod != http.MethodDelete {
		t.Errorf("last method = %q, want DELETE", lastMethod)
	}
	if lastPath != wantPath {
		t.Errorf("DELETE path = %q, want %q", lastPath, wantPath)
	}
}

func TestClient_ListStringParameters_FiltersByPid(t *testing.T) {
	var gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": [{"Pid":"PartnerZ","Id":"A","Value":"1"},{"Pid":"PartnerZ","Id":"B","Value":"2"}]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	params, err := client.ListStringParameters(context.Background(), "PartnerZ")
	if err != nil {
		t.Fatalf("ListStringParameters() error: %v", err)
	}
	if len(params) != 2 {
		t.Fatalf("got %d params, want 2", len(params))
	}
	if gotQuery == "" {
		t.Error("expected a $filter query string scoping the list to the given Pid")
	}
}

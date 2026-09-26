package cloudintegration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A tenant answered PATCH on IntegrationPackages with 501 and accepted a PUT
// of {Id, Name, Description, ShortText} (September 2026).
func TestClient_UpdatePackage(t *testing.T) {
	var body map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/IntegrationPackages('UTILITIES')" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	err := client.UpdatePackage(context.Background(), "UTILITIES", Package{Name: "Utilities", Description: "updated", ShortText: "short"})
	if err != nil {
		t.Fatalf("UpdatePackage() error: %v", err)
	}
	want := map[string]string{"Id": "UTILITIES", "Name": "Utilities", "Description": "updated", "ShortText": "short"}
	if len(body) != len(want) {
		t.Fatalf("body = %v, want %v", body, want)
	}
	for k, v := range want {
		if body[k] != v {
			t.Errorf("body[%s] = %q, want %q", k, body[k], v)
		}
	}
}

// Create without ShortText fails on a tenant, so the body always carries it.
func TestClient_CreatePackage_SendsShortText(t *testing.T) {
	var body map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d":{"Id":"P","Name":"P","Description":"<p>d</p>","ShortText":"s"}}`))
	}))
	defer server.Close()

	if _, err := New(http.DefaultClient, server.URL).CreatePackage(context.Background(), Package{ID: "P", Name: "P", Description: "d", ShortText: "s"}); err != nil {
		t.Fatalf("CreatePackage() error: %v", err)
	}
	if body["ShortText"] != "s" || body["Id"] != "P" {
		t.Errorf("body = %v, want Id and ShortText", body)
	}
}

func TestPlainDescription(t *testing.T) {
	for in, want := range map[string]string{
		"<p>tf-acc probe</p>": "tf-acc probe",
		"<p></p>":             "",
		"<p>a &amp; b</p>":    "a & b",
		"plain":               "plain",
		"<p>a</p><p>b</p>":    "<p>a</p><p>b</p>",
		"<p><b>bold</b></p>":  "<p><b>bold</b></p>",
		"":                    "",
	} {
		if got := PlainDescription(in); got != want {
			t.Errorf("PlainDescription(%q) = %q, want %q", in, got, want)
		}
	}
}

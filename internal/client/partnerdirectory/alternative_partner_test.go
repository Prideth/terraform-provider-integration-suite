package partnerdirectory

import (
	"context"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestEncodeAlternativePartnerKey_MatchesSAPDocumentedExample pins the
// encoding down against a concrete, documented SAP example: the agency
// string "agency1" is confirmed to hex-encode to "6167656e637931" in an
// actual AlternativePartners request URL. If this ever stops matching,
// every AlternativePartner CRUD/import operation would silently address
// the wrong entity.
func TestEncodeAlternativePartnerKey_MatchesSAPDocumentedExample(t *testing.T) {
	hexAgency, _, _ := EncodeAlternativePartnerKey("agency1", "", "")
	if hexAgency != "6167656e637931" {
		t.Errorf("EncodeAlternativePartnerKey(%q) = %q, want %q", "agency1", hexAgency, "6167656e637931")
	}
}

func TestEncodeAlternativePartnerKey_RoundTrips(t *testing.T) {
	cases := []struct {
		name               string
		agency, scheme, id string
	}{
		{"ascii", "Sender_1", "SenderInterface", "Interface_1"},
		{"spaces", "Agency With Spaces", "Scheme Name", "Some Id Value"},
		{"punctuation", "agency!@#$%^&*()", "scheme-._~", "id,;:'\"<>?"},
		{"unicode", "Lieferant_Müller", "Схема", "識別子"},
		{"empty_id_only", "Sender_1", "SenderInterface", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hexAgency, hexScheme, hexID := EncodeAlternativePartnerKey(c.agency, c.scheme, c.id)

			for _, h := range []string{hexAgency, hexScheme, hexID} {
				if _, err := hex.DecodeString(h); err != nil {
					t.Fatalf("encoded value %q is not valid hex: %v", h, err)
				}
			}

			gotAgency, err := DecodeAlternativePartnerKeyComponent(hexAgency)
			if err != nil {
				t.Fatalf("decoding agency: %v", err)
			}
			if gotAgency != c.agency {
				t.Errorf("agency round-trip = %q, want %q", gotAgency, c.agency)
			}

			gotScheme, err := DecodeAlternativePartnerKeyComponent(hexScheme)
			if err != nil {
				t.Fatalf("decoding scheme: %v", err)
			}
			if gotScheme != c.scheme {
				t.Errorf("scheme round-trip = %q, want %q", gotScheme, c.scheme)
			}

			gotID, err := DecodeAlternativePartnerKeyComponent(hexID)
			if err != nil {
				t.Fatalf("decoding id: %v", err)
			}
			if gotID != c.id {
				t.Errorf("id round-trip = %q, want %q", gotID, c.id)
			}
		})
	}
}

func TestEncodeAlternativePartnerKey_EmptyStringEncodesToEmptyHex(t *testing.T) {
	hexAgency, hexScheme, hexID := EncodeAlternativePartnerKey("", "", "")
	if hexAgency != "" || hexScheme != "" || hexID != "" {
		t.Errorf("encoding empty strings = (%q, %q, %q), want all empty", hexAgency, hexScheme, hexID)
	}
}

func TestEncodeAlternativePartnerKey_DistinctInputsNeverCollide(t *testing.T) {
	inputs := []string{"a", "b", "A", "aa", "a ", " a", "a\x00b", "a/b", "a\\b"}
	seen := map[string]string{}
	for _, in := range inputs {
		encoded, _, _ := EncodeAlternativePartnerKey(in, "", "")
		if prior, ok := seen[encoded]; ok && prior != in {
			t.Errorf("inputs %q and %q both encode to %q", prior, in, encoded)
		}
		seen[encoded] = in
	}
}

func TestDecodeAlternativePartnerKeyComponent_RejectsInvalidHex(t *testing.T) {
	if _, err := DecodeAlternativePartnerKeyComponent("not-hex!"); err == nil {
		t.Error("expected an error decoding a non-hex string")
	}
}

func TestClient_GetCreateUpdateDeleteAlternativePartner(t *testing.T) {
	var lastMethod, lastPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path

		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"Agency":"Sender_1","Scheme":"SenderInterface","Id":"Interface_1","Pid":"CS_Scenario_1"}}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"Agency":"Sender_1","Scheme":"SenderInterface","Id":"Interface_1","Pid":"CS_Scenario_1"}}`))
		case http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	ap, err := client.GetAlternativePartner(context.Background(), "Sender_1", "SenderInterface", "Interface_1")
	if err != nil {
		t.Fatalf("GetAlternativePartner() error: %v", err)
	}
	if ap.Pid != "CS_Scenario_1" {
		t.Errorf("Pid = %q, want CS_Scenario_1", ap.Pid)
	}

	wantHexAgency, wantHexScheme, wantHexID := EncodeAlternativePartnerKey("Sender_1", "SenderInterface", "Interface_1")
	wantPath := "/api/v1/AlternativePartners(Hexagency='" + wantHexAgency + "',Hexscheme='" + wantHexScheme + "',Hexid='" + wantHexID + "')"
	if lastPath != wantPath {
		t.Errorf("GET path = %q, want %q", lastPath, wantPath)
	}

	if _, err := client.CreateAlternativePartner(context.Background(), AlternativePartner{
		Agency: "Sender_1", Scheme: "SenderInterface", Id: "Interface_1", Pid: "CS_Scenario_1",
	}); err != nil {
		t.Fatalf("CreateAlternativePartner() error: %v", err)
	}
	if lastPath != "/api/v1/AlternativePartners" {
		t.Errorf("POST path = %q, want /api/v1/AlternativePartners", lastPath)
	}

	if err := client.UpdateAlternativePartner(context.Background(), "Sender_1", "SenderInterface", "Interface_1", "CS_Scenario_2"); err != nil {
		t.Fatalf("UpdateAlternativePartner() error: %v", err)
	}
	if lastMethod != http.MethodPut || lastPath != wantPath {
		t.Errorf("PUT: method = %q path = %q, want PUT %q", lastMethod, lastPath, wantPath)
	}

	if err := client.DeleteAlternativePartner(context.Background(), "Sender_1", "SenderInterface", "Interface_1"); err != nil {
		t.Fatalf("DeleteAlternativePartner() error: %v", err)
	}
	if lastMethod != http.MethodDelete || lastPath != wantPath {
		t.Errorf("DELETE: method = %q path = %q, want DELETE %q", lastMethod, lastPath, wantPath)
	}
}

func TestClient_ListAlternativePartners(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": [{"Agency":"A","Scheme":"S","Id":"I","Pid":"P"}]}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)

	partners, err := client.ListAlternativePartners(context.Background(), "P")
	if err != nil {
		t.Fatalf("ListAlternativePartners() error: %v", err)
	}
	if len(partners) != 1 {
		t.Fatalf("got %d partners, want 1", len(partners))
	}
}

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/securitycontent"
)

func TestKeystoreEntryDataSource_SchemaRequiredComputed(t *testing.T) {
	d := NewKeystoreEntryDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"alias", true, false},
		{"hex_alias", false, true},
		{"key_type", false, true},
		{"key_size", false, true},
		{"valid_not_before", false, true},
		{"valid_not_after", false, true},
	}

	for _, c := range cases {
		attr, ok := resp.Schema.Attributes[c.name]
		if !ok {
			t.Errorf("missing attribute %q", c.name)
			continue
		}
		if attr.IsRequired() != c.required {
			t.Errorf("%s.Required = %v, want %v", c.name, attr.IsRequired(), c.required)
		}
		if attr.IsComputed() != c.computed {
			t.Errorf("%s.Computed = %v, want %v", c.name, attr.IsComputed(), c.computed)
		}
	}
}

func TestKeystoreEntriesDataSource_Schema(t *testing.T) {
	d := NewKeystoreEntriesDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	entriesAttr, ok := resp.Schema.Attributes["entries"]
	if !ok {
		t.Fatal("missing attribute \"entries\"")
	}
	if !entriesAttr.IsComputed() {
		t.Error("entries must be Computed")
	}
}

func TestKeystoreEntryToModel_MapsMetadataFieldsAndDates(t *testing.T) {
	got := keystoreEntryToModel(&securitycontent.KeystoreEntry{
		Hexalias:          "6d79",
		Alias:             "my",
		KeyType:           "RSA",
		KeySize:           2048,
		ValidNotBefore:    "/Date(1570147200000+0000)/",
		ValidNotAfter:     "/Date(1585742400000+0000)/",
		SubjectDN:         "CN=partner.example.invalid",
		FingerprintSha256: "AB:CD",
		Owner:             "SAP",
		Version:           3,
		LastModifiedTime:  "not-a-date-literal",
	})

	checks := map[string][2]string{
		"valid_not_before":   {got.ValidNotBefore.ValueString(), "2019-10-04T00:00:00Z"},
		"valid_not_after":    {got.ValidNotAfter.ValueString(), "2020-04-01T12:00:00Z"},
		"subject_dn":         {got.SubjectDN.ValueString(), "CN=partner.example.invalid"},
		"fingerprint_sha256": {got.FingerprintSha256.ValueString(), "AB:CD"},
		"owner":              {got.Owner.ValueString(), "SAP"},
		"last_modified_time": {got.LastModifiedTime.ValueString(), "not-a-date-literal"},
	}
	for name, c := range checks {
		if c[0] != c[1] {
			t.Errorf("%s = %q, want %q", name, c[0], c[1])
		}
	}
	if got.CertificateVersion.ValueInt64() != 3 {
		t.Errorf("certificate_version = %d, want 3", got.CertificateVersion.ValueInt64())
	}
	if !got.CreatedTime.IsNull() || !got.EllipticCurve.IsNull() {
		t.Error("fields SAP did not return must be null")
	}
}

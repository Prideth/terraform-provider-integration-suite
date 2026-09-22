package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

func partnerBinaryParameterSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewPartnerBinaryParameterResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestPartnerBinaryParameterResource_SchemaRequiredComputed(t *testing.T) {
	s := partnerBinaryParameterSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"partner_id", true, false},
		{"parameter_id", true, false},
		{"content_type", true, false},
		{"content", true, false},
		{"content_hash", true, false},
	}

	for _, c := range cases {
		attr, ok := s.Attributes[c.name]
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

func TestPartnerBinaryParameterResource_ImportState(t *testing.T) {
	r := NewPartnerBinaryParameterResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, partnerBinaryParameterSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "PartnerZ/OrderSchema"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var partnerID, parameterID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("partner_id"), &partnerID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("parameter_id"), &parameterID)...)
	if partnerID.ValueString() != "PartnerZ" {
		t.Errorf("partner_id = %q, want PartnerZ", partnerID.ValueString())
	}
	if parameterID.ValueString() != "OrderSchema" {
		t.Errorf("parameter_id = %q, want OrderSchema", parameterID.ValueString())
	}
}

func TestBinaryParameterToModel_PreservesPreviousContentFields(t *testing.T) {
	previous := partnerBinaryParameterModel{
		Content:     types.StringValue("${path.module}/order.xsd"),
		ContentHash: types.StringValue("deadbeef"),
	}
	bp := &partnerdirectory.BinaryParameter{Pid: "PartnerZ", Id: "OrderSchema", ContentType: "xsd"}

	got := binaryParameterToModel(bp, previous)

	if got.ID.ValueString() != "PartnerZ/OrderSchema" {
		t.Errorf("ID = %q, want PartnerZ/OrderSchema", got.ID.ValueString())
	}
	if got.Content != previous.Content {
		t.Errorf("Content = %v, want preserved from previous (%v)", got.Content, previous.Content)
	}
	if got.ContentHash != previous.ContentHash {
		t.Errorf("ContentHash = %v, want preserved from previous (%v)", got.ContentHash, previous.ContentHash)
	}
}

func writeTempFile(t *testing.T, content []byte) (path, hash string) {
	t.Helper()
	dir := t.TempDir()
	path = filepath.Join(dir, "content.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	sum := sha256.Sum256(content)
	return path, hex.EncodeToString(sum[:])
}

func TestReadAndValidateBinaryParameterContent_Success(t *testing.T) {
	content := []byte("<xsd>schema</xsd>")
	path, hash := writeTempFile(t, content)

	got, err := readAndValidateBinaryParameterContent(path, hash)
	if err != nil {
		t.Fatalf("readAndValidateBinaryParameterContent() error: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestReadAndValidateBinaryParameterContent_HashMismatch(t *testing.T) {
	path, _ := writeTempFile(t, []byte("<xsd>schema</xsd>"))

	if _, err := readAndValidateBinaryParameterContent(path, "wronghash"); err == nil {
		t.Fatal("expected a hash mismatch error")
	}
}

func TestReadAndValidateBinaryParameterContent_RejectsOversizedContent(t *testing.T) {
	oversized := make([]byte, partnerdirectory.MaxBinaryParameterValueBytes+1)
	path, hash := writeTempFile(t, oversized)

	_, err := readAndValidateBinaryParameterContent(path, hash)
	if err == nil {
		t.Fatal("expected an error for content exceeding the 260 KB SAP limit")
	}
}

func TestReadAndValidateBinaryParameterContent_AllowsExactLimit(t *testing.T) {
	exact := make([]byte, partnerdirectory.MaxBinaryParameterValueBytes)
	path, hash := writeTempFile(t, exact)

	if _, err := readAndValidateBinaryParameterContent(path, hash); err != nil {
		t.Errorf("content at exactly the limit should be accepted, got error: %v", err)
	}
}

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

func accessPolicyReferenceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewAccessPolicyReferenceResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestAccessPolicyReferenceResource_SchemaRequiredComputed(t *testing.T) {
	s := accessPolicyReferenceSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"access_policy_id", true, false},
		{"artifact_type", true, false},
		{"attribute", true, false},
		{"operator", true, false},
		{"value", true, false},
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

// TestAccessPolicyReferenceResource_EveryMutableFieldRequiresReplace pins
// down the deliberate design decision behind this resource: SAP does not
// document an in-place update for an artifact reference's match condition,
// so every field that identifies or defines the reference forces
// replacement, and Update (see resource_access_policy_reference.go) should
// therefore never actually be invoked by Terraform.
func TestAccessPolicyReferenceResource_EveryMutableFieldRequiresReplace(t *testing.T) {
	s := accessPolicyReferenceSchema(t).Schema

	fields := []string{"access_policy_id", "artifact_type", "attribute", "operator", "value"}
	for _, name := range fields {
		attr, ok := s.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("attribute %q is not a StringAttribute", name)
		}

		var mods []interface {
			Description(context.Context) string
		}
		for _, m := range attr.PlanModifiers {
			mods = append(mods, m)
		}

		if !hasRequiresReplace(mods) {
			t.Errorf("%s has no RequiresReplace plan modifier", name)
		}
	}
}

func TestAccessPolicyReferenceResource_Update_AlwaysErrors(t *testing.T) {
	r := NewAccessPolicyReferenceResource()

	var resp resource.UpdateResponse
	r.Update(context.Background(), resource.UpdateRequest{}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Errorf("Update() did not produce a diagnostic; every field is RequiresReplace so Update should be unreachable")
	}
}

func TestAccessPolicyReferenceResource_ImportState(t *testing.T) {
	r := NewAccessPolicyReferenceResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, accessPolicyReferenceSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "1/ref-1"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var accessPolicyID, id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("access_policy_id"), &accessPolicyID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRootID(), &id)...)
	if accessPolicyID.ValueString() != "1" {
		t.Errorf("access_policy_id = %q, want %q", accessPolicyID.ValueString(), "1")
	}
	if id.ValueString() != "1/ref-1" {
		t.Errorf("id = %q, want %q", id.ValueString(), "1/ref-1")
	}
}

func TestAccessPolicyReferenceResource_ImportState_InvalidID(t *testing.T) {
	r := NewAccessPolicyReferenceResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, accessPolicyReferenceSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "not-composite"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Errorf("ImportState() with a non-composite ID should produce a diagnostic")
	}
}

func TestReferenceToModel(t *testing.T) {
	ref := &cloudintegration.AccessPolicyReference{
		ID:           "ref-1",
		ArtifactType: "IntegrationFlow",
		Attribute:    "Name",
		Operator:     "MATCHES",
		Value:        "UTIL_.*",
	}

	got := referenceToModel("1", ref)

	if got.ID.ValueString() != "1/ref-1" {
		t.Errorf("ID = %q, want %q", got.ID.ValueString(), "1/ref-1")
	}
	if got.AccessPolicyID.ValueString() != "1" {
		t.Errorf("AccessPolicyID = %q, want %q", got.AccessPolicyID.ValueString(), "1")
	}
	if got.Value.ValueString() != "UTIL_.*" {
		t.Errorf("Value = %q, want %q", got.Value.ValueString(), "UTIL_.*")
	}
}

func TestReferenceIDFrom(t *testing.T) {
	if got := referenceIDFrom("1/ref-1"); got != "ref-1" {
		t.Errorf("referenceIDFrom(%q) = %q, want %q", "1/ref-1", got, "ref-1")
	}
}

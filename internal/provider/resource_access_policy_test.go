package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

func accessPolicySchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewAccessPolicyResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestAccessPolicyResource_SchemaRequiredOptionalComputed(t *testing.T) {
	s := accessPolicySchema(t).Schema

	cases := []struct {
		name               string
		required, optional bool
		computed           bool
	}{
		{"id", false, false, true},
		{"role_name", true, false, false},
		{"description", false, true, false},
	}

	if len(s.Attributes) != len(cases) {
		t.Errorf("schema has %d attributes, want %d", len(s.Attributes), len(cases))
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
		if attr.IsOptional() != c.optional {
			t.Errorf("%s.Optional = %v, want %v", c.name, attr.IsOptional(), c.optional)
		}
		if attr.IsComputed() != c.computed {
			t.Errorf("%s.Computed = %v, want %v", c.name, attr.IsComputed(), c.computed)
		}
	}
}

// reconciliation_status was removed because the AccessPolicies entity has no
// such property; runtime replication lives in AccessPolicyRuntimeAssignments.
func TestAccessPolicyResource_NoReconciliationStatus(t *testing.T) {
	if _, ok := accessPolicySchema(t).Schema.Attributes["reconciliation_status"]; ok {
		t.Error("reconciliation_status must not be part of the schema")
	}
}

func TestAccessPolicyResource_RoleNameRequiresReplace(t *testing.T) {
	s := accessPolicySchema(t).Schema

	replaceExpected := map[string]bool{
		"role_name":   true,
		"description": false,
	}

	for name, wantReplace := range replaceExpected {
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

		if got := hasRequiresReplace(mods); got != wantReplace {
			t.Errorf("%s: RequiresReplace present = %v, want %v", name, got, wantReplace)
		}
	}
}

func TestAccessPolicyResource_ImportState(t *testing.T) {
	r := NewAccessPolicyResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, accessPolicySchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "1901"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRootID(), &id)...)
	if id.ValueString() != "1901" {
		t.Errorf("id = %q, want %q", id.ValueString(), "1901")
	}
}

func TestAccessPolicyResource_ImportState_RejectsNonNumericID(t *testing.T) {
	r := NewAccessPolicyResource().(resource.ResourceWithImportState)

	for _, id := range []string{"UTILITIES_ARCHITECT", "", "1.5"} {
		resp := &resource.ImportStateResponse{State: newTestState(t, accessPolicySchema(t).Schema)}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("ImportState(%q) should produce a diagnostic", id)
		}
	}
}

func TestAccessPolicyToModel(t *testing.T) {
	policy := &cloudintegration.AccessPolicy{
		ID:       "1901",
		RoleName: "UTILITIES_ARCHITECT",
	}

	got := accessPolicyToModel(policy)

	if got.ID.ValueString() != "1901" {
		t.Errorf("ID = %q, want %q", got.ID.ValueString(), "1901")
	}
	if !got.Description.IsNull() {
		t.Errorf("Description = %v, want null when SAP reports an empty description", got.Description)
	}

	policy.Description = "desc"
	got = accessPolicyToModel(policy)
	if got.Description.ValueString() != "desc" {
		t.Errorf("Description = %q, want %q", got.Description.ValueString(), "desc")
	}
}

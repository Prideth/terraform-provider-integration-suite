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
		{"reconciliation_status", false, false, true},
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

// TestAccessPolicyResource_RoleNameRequiresReplace pins down that role_name
// forces replacement: SAP does not document renaming a policy's role, so a
// changed role_name must never be sent through Update.
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
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "1"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRootID(), &id)...)
	if id.ValueString() != "1" {
		t.Errorf("id = %q, want %q", id.ValueString(), "1")
	}
}

func TestAccessPolicyToModel(t *testing.T) {
	policy := &cloudintegration.AccessPolicy{
		ID:       "1",
		RoleName: "UTILITIES_ARCHITECT",
	}

	got := accessPolicyToModel(policy)

	if got.ID.ValueString() != "1" {
		t.Errorf("ID = %q, want %q", got.ID.ValueString(), "1")
	}
	if !got.Description.IsNull() {
		t.Errorf("Description = %v, want null when SAP reports no description", got.Description)
	}
	if !got.ReconciliationStatus.IsNull() {
		t.Errorf("ReconciliationStatus = %v, want null when SAP reports none", got.ReconciliationStatus)
	}

	policy.Description = "desc"
	policy.ReconciliationStatus = "SUCCESS"
	got = accessPolicyToModel(policy)
	if got.Description.ValueString() != "desc" {
		t.Errorf("Description = %q, want %q", got.Description.ValueString(), "desc")
	}
	if got.ReconciliationStatus.ValueString() != "SUCCESS" {
		t.Errorf("ReconciliationStatus = %q, want %q", got.ReconciliationStatus.ValueString(), "SUCCESS")
	}
}

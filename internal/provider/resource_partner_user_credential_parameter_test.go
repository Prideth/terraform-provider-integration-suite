package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func partnerUserCredentialParameterSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewPartnerUserCredentialParameterResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestPartnerUserCredentialParameterResource_SchemaRequiredComputed(t *testing.T) {
	s := partnerUserCredentialParameterSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"partner_id", true, false},
		{"parameter_id", true, false},
		{"user", true, false},
		{"password_wo", true, false},
		{"password_wo_version", true, false},
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

// TestPartnerUserCredentialParameterResource_PasswordIsWriteOnlyAndSensitive
// is the safety property this whole resource depends on: password_wo must
// never be persisted to plan or state.
func TestPartnerUserCredentialParameterResource_PasswordIsWriteOnlyAndSensitive(t *testing.T) {
	s := partnerUserCredentialParameterSchema(t).Schema

	attr, ok := s.Attributes["password_wo"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("password_wo is not a StringAttribute")
	}
	if !attr.IsWriteOnly() {
		t.Error("password_wo must be WriteOnly")
	}
	if !attr.IsSensitive() {
		t.Error("password_wo must be Sensitive")
	}
}

// TestPartnerUserCredentialParameterResource_EveryFieldForcesReplace pins
// down that no in-place update path exists: identity, the communication
// user, and the version marker that signals a password rotation all force
// replacement, since no confirmed public API exists for updating a stored
// credential's password.
func TestPartnerUserCredentialParameterResource_EveryFieldForcesReplace(t *testing.T) {
	s := partnerUserCredentialParameterSchema(t).Schema

	fields := []string{"partner_id", "parameter_id", "user", "password_wo_version"}
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

func TestPartnerUserCredentialParameterResource_Update_AlwaysErrors(t *testing.T) {
	r := NewPartnerUserCredentialParameterResource()

	var resp resource.UpdateResponse
	r.Update(context.Background(), resource.UpdateRequest{}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("Update() did not produce a diagnostic; every field is RequiresReplace so Update should be unreachable")
	}
}

func TestPartnerUserCredentialParameterResource_ImportState(t *testing.T) {
	r := NewPartnerUserCredentialParameterResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, partnerUserCredentialParameterSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "Receiver_1/USER"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var partnerID, parameterID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("partner_id"), &partnerID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("parameter_id"), &parameterID)...)
	if partnerID.ValueString() != "Receiver_1" {
		t.Errorf("partner_id = %q, want Receiver_1", partnerID.ValueString())
	}
	if parameterID.ValueString() != "USER" {
		t.Errorf("parameter_id = %q, want USER", parameterID.ValueString())
	}
}

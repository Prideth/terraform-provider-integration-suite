package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

func alternativePartnerSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewAlternativePartnerResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestAlternativePartnerResource_SchemaRequiredComputed(t *testing.T) {
	s := alternativePartnerSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"agency", true, false},
		{"scheme", true, false},
		{"external_id", true, false},
		{"partner_id", true, false},
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

// TestAlternativePartnerResource_IdentityFieldsRequireReplace_PartnerIDDoesNot
// pins down the deliberate design: the external identity (agency, scheme,
// external_id) is immutable, but partner_id is mutable in place via PUT.
func TestAlternativePartnerResource_IdentityFieldsRequireReplace_PartnerIDDoesNot(t *testing.T) {
	s := alternativePartnerSchema(t).Schema

	replaceExpected := map[string]bool{
		"agency":      true,
		"scheme":      true,
		"external_id": true,
		"partner_id":  false,
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

func TestAlternativePartnerResource_ImportState(t *testing.T) {
	r := NewAlternativePartnerResource().(resource.ResourceWithImportState)

	id := alternativePartnerID("Sender_1", "SenderInterface", "Interface_1")

	resp := &resource.ImportStateResponse{State: newTestState(t, alternativePartnerSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var agency, scheme, externalID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("agency"), &agency)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("scheme"), &scheme)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("external_id"), &externalID)...)

	if agency.ValueString() != "Sender_1" {
		t.Errorf("agency = %q, want Sender_1", agency.ValueString())
	}
	if scheme.ValueString() != "SenderInterface" {
		t.Errorf("scheme = %q, want SenderInterface", scheme.ValueString())
	}
	if externalID.ValueString() != "Interface_1" {
		t.Errorf("external_id = %q, want Interface_1", externalID.ValueString())
	}
}

func TestAlternativePartnerResource_ImportState_InvalidID(t *testing.T) {
	r := NewAlternativePartnerResource().(resource.ResourceWithImportState)

	cases := []string{"not-three-parts", "a/b", "a/b/c/d", "zz/61/62"}
	for _, id := range cases {
		resp := &resource.ImportStateResponse{State: newTestState(t, alternativePartnerSchema(t).Schema)}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("ImportState(%q) should have produced a diagnostic", id)
		}
	}
}

func TestAlternativePartnerToModel(t *testing.T) {
	ap := &partnerdirectory.AlternativePartner{Agency: "Sender_1", Scheme: "SenderInterface", Id: "Interface_1", Pid: "CS_Scenario_1"}
	got := alternativePartnerToModel(ap)

	wantID := alternativePartnerID("Sender_1", "SenderInterface", "Interface_1")
	if got.ID.ValueString() != wantID {
		t.Errorf("ID = %q, want %q", got.ID.ValueString(), wantID)
	}
	if got.PartnerID.ValueString() != "CS_Scenario_1" {
		t.Errorf("PartnerID = %q, want CS_Scenario_1", got.PartnerID.ValueString())
	}
}

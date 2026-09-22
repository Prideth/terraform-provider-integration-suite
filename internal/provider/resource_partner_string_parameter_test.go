package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

func partnerStringParameterSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewPartnerStringParameterResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestPartnerStringParameterResource_SchemaRequiredComputed(t *testing.T) {
	s := partnerStringParameterSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"partner_id", true, false},
		{"parameter_id", true, false},
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

func TestPartnerStringParameterResource_ImportState(t *testing.T) {
	r := NewPartnerStringParameterResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, partnerStringParameterSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "PartnerZ/ReceiverAddress"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var partnerID, parameterID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("partner_id"), &partnerID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("parameter_id"), &parameterID)...)
	if partnerID.ValueString() != "PartnerZ" {
		t.Errorf("partner_id = %q, want PartnerZ", partnerID.ValueString())
	}
	if parameterID.ValueString() != "ReceiverAddress" {
		t.Errorf("parameter_id = %q, want ReceiverAddress", parameterID.ValueString())
	}
}

func TestStringParameterToModel(t *testing.T) {
	sp := &partnerdirectory.StringParameter{Pid: "PartnerZ", Id: "ReceiverAddress", Value: "https://receiver.example.com"}
	got := stringParameterToModel(sp)
	if got.ID.ValueString() != "PartnerZ/ReceiverAddress" {
		t.Errorf("ID = %q, want PartnerZ/ReceiverAddress", got.ID.ValueString())
	}
	if got.Value.ValueString() != "https://receiver.example.com" {
		t.Errorf("Value = %q, want the receiver URL", got.Value.ValueString())
	}
}

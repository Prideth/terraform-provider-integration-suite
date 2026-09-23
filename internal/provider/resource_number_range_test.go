package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

func numberRangeSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewNumberRangeResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestNumberRangeResource_SchemaRequiredComputed(t *testing.T) {
	s := numberRangeSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"name", true, false},
		{"min_value", true, false},
		{"max_value", true, false},
		{"description", false, true},
		{"rotate", true, false},
		{"field_length", false, true},
		{"current_value_wo", true, false},
		{"current_value_wo_version", true, false},
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

	nameAttr, ok := s.Attributes["name"].(schema.StringAttribute)
	if !ok || len(nameAttr.PlanModifiers) == 0 {
		t.Error("name must carry a RequiresReplace plan modifier")
	}

	currentValueAttr, ok := s.Attributes["current_value_wo"].(schema.StringAttribute)
	if !ok || !currentValueAttr.WriteOnly {
		t.Error("current_value_wo must be WriteOnly")
	}
}

func TestNumberRangeResource_ValidateConfig_RejectsMinGreaterThanMax(t *testing.T) {
	r := NewNumberRangeResource().(resource.ResourceWithValidateConfig)
	s := numberRangeSchema(t).Schema
	ctx := context.Background()

	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, nil),
		"name":                     tftypes.NewValue(tftypes.String, "MyRange"),
		"min_value":                tftypes.NewValue(tftypes.String, "9999"),
		"max_value":                tftypes.NewValue(tftypes.String, "0"),
		"description":              tftypes.NewValue(tftypes.String, nil),
		"rotate":                   tftypes.NewValue(tftypes.Bool, true),
		"field_length":             tftypes.NewValue(tftypes.String, nil),
		"current_value_wo":         tftypes.NewValue(tftypes.String, nil),
		"current_value_wo_version": tftypes.NewValue(tftypes.String, "v1"),
	})

	req := resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: s, Raw: raw}}
	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ValidateConfig() did not reject min_value > max_value")
	}
}

func TestNumberRangeResource_ValidateConfig_AcceptsValidRange(t *testing.T) {
	r := NewNumberRangeResource().(resource.ResourceWithValidateConfig)
	s := numberRangeSchema(t).Schema
	ctx := context.Background()

	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, nil),
		"name":                     tftypes.NewValue(tftypes.String, "MyRange"),
		"min_value":                tftypes.NewValue(tftypes.String, "0"),
		"max_value":                tftypes.NewValue(tftypes.String, "9999"),
		"description":              tftypes.NewValue(tftypes.String, nil),
		"rotate":                   tftypes.NewValue(tftypes.Bool, true),
		"field_length":             tftypes.NewValue(tftypes.String, nil),
		"current_value_wo":         tftypes.NewValue(tftypes.String, nil),
		"current_value_wo_version": tftypes.NewValue(tftypes.String, "v1"),
	})

	req := resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: s, Raw: raw}}
	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ValidateConfig() unexpectedly produced diagnostics: %v", resp.Diagnostics)
	}
}

func TestNumberRangeResource_Delete_AlwaysErrors(t *testing.T) {
	r := NewNumberRangeResource()

	var resp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("Delete() did not produce a diagnostic; SAP documents no delete operation for this entity")
	}
}

func TestNumberRangeResource_ImportState_AlwaysErrors(t *testing.T) {
	r := NewNumberRangeResource().(resource.ResourceWithImportState)

	var resp resource.ImportStateResponse
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "MyRange"}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("ImportState() did not produce a diagnostic; SAP documents no GET for this entity, so import cannot populate state from the tenant")
	}
}

func newNumberRangeTestState(t *testing.T, s schema.Schema) tfsdk.State {
	t.Helper()
	return tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}
}

// TestNumberRangeResource_Update_OmitsCurrentValueWhenVersionUnchanged is the
// mandatory regression test for the counter-preservation contract: a
// practitioner changing only "description" (current_value_wo_version left
// untouched) must produce a PUT request whose body contains no CurrentValue
// property at all. This is this resource's version of the "never rewind the
// counter" requirement, adapted to a client with no GET: instead of
// re-fetching the live value before Update, it simply never sends one
// unless explicitly asked to via a version bump.
func TestNumberRangeResource_Update_OmitsCurrentValueWhenVersionUnchanged(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	r := &numberRangeResource{client: cloudintegration.New(http.DefaultClient, server.URL)}
	s := numberRangeSchema(t).Schema
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	stateRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "MyRange"),
		"name":                     tftypes.NewValue(tftypes.String, "MyRange"),
		"min_value":                tftypes.NewValue(tftypes.String, "0"),
		"max_value":                tftypes.NewValue(tftypes.String, "9999"),
		"description":              tftypes.NewValue(tftypes.String, "old description"),
		"rotate":                   tftypes.NewValue(tftypes.Bool, true),
		"field_length":             tftypes.NewValue(tftypes.String, "4"),
		"current_value_wo":         tftypes.NewValue(tftypes.String, nil),
		"current_value_wo_version": tftypes.NewValue(tftypes.String, "v1"),
	})
	planRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "MyRange"),
		"name":                     tftypes.NewValue(tftypes.String, "MyRange"),
		"min_value":                tftypes.NewValue(tftypes.String, "0"),
		"max_value":                tftypes.NewValue(tftypes.String, "9999"),
		"description":              tftypes.NewValue(tftypes.String, "new description"),
		"rotate":                   tftypes.NewValue(tftypes.Bool, true),
		"field_length":             tftypes.NewValue(tftypes.String, "4"),
		"current_value_wo":         tftypes.NewValue(tftypes.String, nil),
		"current_value_wo_version": tftypes.NewValue(tftypes.String, "v1"),
	})
	configRaw := planRaw

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: s, Raw: stateRaw},
		Plan:   tfsdk.Plan{Schema: s, Raw: planRaw},
		Config: tfsdk.Config{Schema: s, Raw: configRaw},
	}
	resp := &resource.UpdateResponse{State: newNumberRangeTestState(t, s)}

	r.Update(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update() produced diagnostics: %v", resp.Diagnostics)
	}

	if containsCurrentValue(gotBody) {
		t.Errorf("PUT body unexpectedly contains CurrentValue when current_value_wo_version did not change: %s", gotBody)
	}
}

// TestNumberRangeResource_Update_SendsCurrentValueWhenVersionChanges proves
// the deliberate-reset path: bumping current_value_wo_version does cause
// CurrentValue to be sent, mirroring this provider's credential-rotation
// pattern.
func TestNumberRangeResource_Update_SendsCurrentValueWhenVersionChanges(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	r := &numberRangeResource{client: cloudintegration.New(http.DefaultClient, server.URL)}
	s := numberRangeSchema(t).Schema
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	stateRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "MyRange"),
		"name":                     tftypes.NewValue(tftypes.String, "MyRange"),
		"min_value":                tftypes.NewValue(tftypes.String, "0"),
		"max_value":                tftypes.NewValue(tftypes.String, "9999"),
		"description":              tftypes.NewValue(tftypes.String, "d"),
		"rotate":                   tftypes.NewValue(tftypes.Bool, true),
		"field_length":             tftypes.NewValue(tftypes.String, "4"),
		"current_value_wo":         tftypes.NewValue(tftypes.String, nil),
		"current_value_wo_version": tftypes.NewValue(tftypes.String, "v1"),
	})
	planRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "MyRange"),
		"name":                     tftypes.NewValue(tftypes.String, "MyRange"),
		"min_value":                tftypes.NewValue(tftypes.String, "0"),
		"max_value":                tftypes.NewValue(tftypes.String, "9999"),
		"description":              tftypes.NewValue(tftypes.String, "d"),
		"rotate":                   tftypes.NewValue(tftypes.Bool, true),
		"field_length":             tftypes.NewValue(tftypes.String, "4"),
		"current_value_wo":         tftypes.NewValue(tftypes.String, nil),
		"current_value_wo_version": tftypes.NewValue(tftypes.String, "v2"),
	})
	configRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, nil),
		"name":                     tftypes.NewValue(tftypes.String, "MyRange"),
		"min_value":                tftypes.NewValue(tftypes.String, "0"),
		"max_value":                tftypes.NewValue(tftypes.String, "9999"),
		"description":              tftypes.NewValue(tftypes.String, "d"),
		"rotate":                   tftypes.NewValue(tftypes.Bool, true),
		"field_length":             tftypes.NewValue(tftypes.String, "4"),
		"current_value_wo":         tftypes.NewValue(tftypes.String, "500"),
		"current_value_wo_version": tftypes.NewValue(tftypes.String, "v2"),
	})

	req := resource.UpdateRequest{
		State:  tfsdk.State{Schema: s, Raw: stateRaw},
		Plan:   tfsdk.Plan{Schema: s, Raw: planRaw},
		Config: tfsdk.Config{Schema: s, Raw: configRaw},
	}
	resp := &resource.UpdateResponse{State: newNumberRangeTestState(t, s)}

	r.Update(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update() produced diagnostics: %v", resp.Diagnostics)
	}

	if !containsCurrentValue(gotBody) {
		t.Errorf("PUT body should contain CurrentValue when current_value_wo_version changed: %s", gotBody)
	}
}

func containsCurrentValue(body []byte) bool {
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(body, &decoded); err != nil {
		return false
	}
	_, ok := decoded["CurrentValue"]
	return ok
}

// TestNumberRangeResource_Read_IsNoOp proves Read never mutates what is
// already in state: it copies state straight through to the response
// without contacting SAP (there is no client on the resource in this test,
// so any attempt to call SAP would nil-panic).
func TestNumberRangeResource_Read_IsNoOp(t *testing.T) {
	r := &numberRangeResource{}
	s := numberRangeSchema(t).Schema
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	stateRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                       tftypes.NewValue(tftypes.String, "MyRange"),
		"name":                     tftypes.NewValue(tftypes.String, "MyRange"),
		"min_value":                tftypes.NewValue(tftypes.String, "0"),
		"max_value":                tftypes.NewValue(tftypes.String, "9999"),
		"description":              tftypes.NewValue(tftypes.String, "d"),
		"rotate":                   tftypes.NewValue(tftypes.Bool, true),
		"field_length":             tftypes.NewValue(tftypes.String, "4"),
		"current_value_wo":         tftypes.NewValue(tftypes.String, nil),
		"current_value_wo_version": tftypes.NewValue(tftypes.String, "v1"),
	})

	req := resource.ReadRequest{State: tfsdk.State{Schema: s, Raw: stateRaw}}
	resp := &resource.ReadResponse{State: newNumberRangeTestState(t, s)}

	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", resp.Diagnostics)
	}

	var got numberRangeModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if got.Description.ValueString() != "d" {
		t.Errorf("Read() changed description to %q, want unchanged %q", got.Description.ValueString(), "d")
	}
	if got.ID.ValueString() != "MyRange" {
		t.Errorf("Read() changed id to %q, want unchanged %q", got.ID.ValueString(), "MyRange")
	}
}

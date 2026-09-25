package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

func accessPolicyRuntimeAssignmentsSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	var resp datasource.SchemaResponse
	NewAccessPolicyRuntimeAssignmentsDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestAccessPolicyRuntimeAssignmentsDataSource_Read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want := "/api/v1/AccessPolicies(1901L)/AccessPolicyRuntimeAssignments"; r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		_, _ = w.Write([]byte(`{"d": {"results": [
			{"Id": "8", "RuntimeLocationId": "eic-plant-a", "TransferStatus": "PENDING", "TransferErrors": null},
			{"Id": "7", "RuntimeLocationId": "cloudintegration", "TransferStatus": "SUCCESS", "TransferErrors": "", "StatusUpdatedAt": "/Date(1767225600000)/"}
		]}}`))
	}))
	defer server.Close()

	ctx := context.Background()
	d := NewAccessPolicyRuntimeAssignmentsDataSource().(*accessPolicyRuntimeAssignmentsDataSource)
	d.client = cloudintegration.New(http.DefaultClient, server.URL)

	s := accessPolicyRuntimeAssignmentsSchema(t).Schema
	config := newFeatureDataSourceConfig(t, s, map[string]tftypes.Value{
		"access_policy_id": tftypes.NewValue(tftypes.String, "1901"),
	})
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}}
	d.Read(ctx, datasource.ReadRequest{Config: config}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", resp.Diagnostics)
	}

	var state accessPolicyRuntimeAssignmentsModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if len(state.Assignments) != 2 {
		t.Fatalf("len(assignments) = %d, want 2", len(state.Assignments))
	}

	first, second := state.Assignments[0], state.Assignments[1]
	if first.RuntimeLocationID.ValueString() != "cloudintegration" {
		t.Errorf("assignments not sorted by runtime_location_id: first = %q", first.RuntimeLocationID.ValueString())
	}
	if first.StatusUpdatedAt.ValueString() != "2026-01-01T00:00:00Z" {
		t.Errorf("status_updated_at = %q, want RFC 3339", first.StatusUpdatedAt.ValueString())
	}
	if !first.TransferErrors.IsNull() || !second.TransferErrors.IsNull() {
		t.Error("empty or null transfer_errors must be null")
	}
	if second.TransferStatus.ValueString() != "PENDING" || !second.StatusUpdatedAt.IsNull() {
		t.Errorf("second assignment = %+v", second)
	}
}

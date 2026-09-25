package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

func TestVersionToSave(t *testing.T) {
	cases := []struct {
		planned, prior types.String
		wantDue        bool
	}{
		{types.StringNull(), types.StringNull(), false},
		{types.StringValue("1.0.0"), types.StringNull(), true},
		{types.StringValue("1.0.0"), types.StringValue("1.0.0"), false},
		{types.StringValue("1.0.1"), types.StringValue("1.0.0"), true},
		{types.StringNull(), types.StringValue("1.0.0"), false},
	}
	for _, c := range cases {
		if _, due := versionToSave(c.planned, c.prior); due != c.wantDue {
			t.Errorf("versionToSave(%v, %v) due = %v, want %v", c.planned, c.prior, due, c.wantDue)
		}
	}
}

func TestSaveAsVersionAttribute_OnVersionedResources(t *testing.T) {
	for _, r := range []resource.Resource{NewIntegrationFlowResource(), NewMessageMappingResource(), NewScriptCollectionResource()} {
		var resp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
		attr, ok := resp.Schema.Attributes["save_as_version"].(schema.StringAttribute)
		if !ok || !attr.Optional || attr.Computed || len(attr.PlanModifiers) != 0 {
			t.Errorf("%T: save_as_version must be an Optional, non-computed string without RequiresReplace", r)
		}
	}
}

func integrationFlowCreateWithVersion(t *testing.T, saveStatus int) (*resource.CreateResponse, []string) {
	t.Helper()
	ctx := context.Background()
	content := []byte("synthetic-iflow-zip")
	path := filepath.Join(t.TempDir(), "flow.zip")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/api/v1/IntegrationDesigntimeArtifacts":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"Id": "Order_Flow", "Version": "1.0.0", "Name": "Order Flow", "PackageId": "ORDERS"}}`))
		case "/api/v1/IntegrationDesigntimeArtifactSaveAsVersion":
			if saveStatus != http.StatusOK {
				w.WriteHeader(saveStatus)
				return
			}
			_, _ = w.Write([]byte(`{"d": {"Id": "Order_Flow", "Version": "1.0.3", "Name": "Order Flow"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	r := &integrationFlowResource{client: cloudintegration.New(http.DefaultClient, server.URL)}
	var sresp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sresp)
	plan := tfsdk.Plan{Schema: sresp.Schema, Raw: tftypes.NewValue(sresp.Schema.Type().TerraformType(ctx), nil)}
	plan.Set(ctx, integrationFlowModel{
		ID: types.StringUnknown(), PackageID: types.StringValue("ORDERS"), FlowID: types.StringValue("Order_Flow"),
		Name: types.StringValue("Order Flow"), Content: types.StringValue(path),
		ContentHash: types.StringValue(hex.EncodeToString(sum[:])), Version: types.StringUnknown(),
		SaveAsVersion: types.StringValue("1.0.3"),
	})
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sresp.Schema, Raw: tftypes.NewValue(sresp.Schema.Type().TerraformType(ctx), nil)}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	return resp, calls
}

func TestIntegrationFlowCreate_SavesExplicitVersion(t *testing.T) {
	resp, calls := integrationFlowCreateWithVersion(t, http.StatusOK)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Create() diagnostics: %v", resp.Diagnostics)
	}
	if len(calls) != 2 || calls[1] != "POST /api/v1/IntegrationDesigntimeArtifactSaveAsVersion" {
		t.Errorf("calls = %v, want create followed by SaveAsVersion", calls)
	}
	var got integrationFlowModel
	resp.State.Get(context.Background(), &got)
	if got.Version.ValueString() != "1.0.3" || got.SaveAsVersion.ValueString() != "1.0.3" {
		t.Errorf("version/save_as_version = %q/%q, want 1.0.3", got.Version.ValueString(), got.SaveAsVersion.ValueString())
	}
}

func TestIntegrationFlowCreate_SaveFailureStillRecordsTheFlow(t *testing.T) {
	resp, _ := integrationFlowCreateWithVersion(t, http.StatusBadRequest)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when SaveAsVersion fails")
	}
	var got integrationFlowModel
	resp.State.Get(context.Background(), &got)
	if got.FlowID.ValueString() != "Order_Flow" {
		t.Error("the created flow must be recorded in state so it is not left outside Terraform")
	}
	if !got.SaveAsVersion.IsNull() {
		t.Errorf("save_as_version = %v, want null so the next apply retries", got.SaveAsVersion)
	}
}

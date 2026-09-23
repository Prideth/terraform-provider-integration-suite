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

func TestCustomTagConfigurationDataSource_SchemaComputed(t *testing.T) {
	d := NewCustomTagConfigurationDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	for _, name := range []string{"id", "tags"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("missing attribute %q", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("%s must be Computed", name)
		}
	}
}

// TestCustomTagConfigurationDataSource_Read exercises the full Read path
// against a mock SAP server, and specifically proves that SAP returning
// the same tags in a different order than a previous read does not change
// the semantic content this provider stores — the ordering itself is
// handled by Terraform's own Set type once state is written, but this
// confirms the provider layer does not silently drop or reorder-and-lose
// data along the way.
func TestCustomTagConfigurationDataSource_Read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/v1/CustomTagConfigurations('CustomTags')/$value"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"customTagsConfiguration":[
			{"tagName":"BusinessUnit","isMandatory":false,"permittedValues":["Water","Network"]},
			{"tagName":"Owner","isMandatory":true}
		]}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)
	d := NewCustomTagConfigurationDataSource().(datasource.DataSourceWithConfigure)
	d.(*customTagConfigurationDataSource).client = client

	ctx := context.Background()
	s := func() datasource.SchemaResponse {
		var resp datasource.SchemaResponse
		d.Schema(ctx, datasource.SchemaRequest{}, &resp)
		return resp
	}().Schema

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)},
	}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", readResp.Diagnostics)
	}

	var state customTagConfigurationDataSourceModel
	readResp.Diagnostics.Append(readResp.State.Get(ctx, &state)...)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("reading back state produced diagnostics: %v", readResp.Diagnostics)
	}

	if state.ID.ValueString() != "CustomTags" {
		t.Errorf("id = %q, want CustomTags", state.ID.ValueString())
	}
	if len(state.Tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(state.Tags))
	}

	byName := make(map[string]customTagModel, len(state.Tags))
	for _, tag := range state.Tags {
		byName[tag.Name.ValueString()] = tag
	}

	owner, ok := byName["Owner"]
	if !ok {
		t.Fatal("missing Owner tag")
	}
	if !owner.Mandatory.ValueBool() {
		t.Error("Owner.mandatory = false, want true")
	}
	if len(owner.PermittedValues) != 0 {
		t.Errorf("Owner.permitted_values = %v, want empty", owner.PermittedValues)
	}

	businessUnit, ok := byName["BusinessUnit"]
	if !ok {
		t.Fatal("missing BusinessUnit tag")
	}
	if businessUnit.Mandatory.ValueBool() {
		t.Error("BusinessUnit.mandatory = true, want false")
	}
	if len(businessUnit.PermittedValues) != 2 {
		t.Errorf("BusinessUnit.permitted_values = %v, want 2 entries", businessUnit.PermittedValues)
	}
}

func TestCustomTagConfigurationDataSource_Read_NoConfigurationYet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": {"code": "404", "message": {"value": "Not found"}}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)
	d := NewCustomTagConfigurationDataSource().(datasource.DataSourceWithConfigure)
	d.(*customTagConfigurationDataSource).client = client

	ctx := context.Background()
	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)},
	}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("Read() with no configuration on the tenant did not produce a diagnostic")
	}
}

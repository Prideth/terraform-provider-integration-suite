package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apimanagementclassic"
)

func TestAPIProductResource_Configure_RequiresAPIManagementClient(t *testing.T) {
	r := NewAPIProductResource().(resource.ResourceWithConfigure)
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: &Data{}}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a configuration error when APIManagementClassicHTTPClient is nil")
	}
}

func TestAPIProductResource_CreateReadDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"name": "SampleProduct", "version": "1", "title": "SampleProduct", "scope": "", "status_code": "PUBLISHED", "isPublished": false, "isRestricted": false}}`))
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"name": "SampleProduct", "version": "1", "title": "SampleProduct", "scope": "", "status_code": "PUBLISHED", "isPublished": false, "isRestricted": false}}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	r := &apiProductResource{client: apimanagementclassic.New(http.DefaultClient, server.URL)}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	additionalPropsType := objType.AttributeTypes["additional_properties"].(tftypes.Set)
	apiProxyNamesType := objType.AttributeTypes["api_proxy_names"].(tftypes.List)

	planRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, nil),
		"name":                  tftypes.NewValue(tftypes.String, "SampleProduct"),
		"version":               tftypes.NewValue(tftypes.String, "1"),
		"title":                 tftypes.NewValue(tftypes.String, "SampleProduct"),
		"description":           tftypes.NewValue(tftypes.String, nil),
		"scope":                 tftypes.NewValue(tftypes.String, ""),
		"status_code":           tftypes.NewValue(tftypes.String, "PUBLISHED"),
		"is_published":          tftypes.NewValue(tftypes.Bool, false),
		"is_restricted":         tftypes.NewValue(tftypes.Bool, false),
		"quota_count":           tftypes.NewValue(tftypes.Number, nil),
		"quota_interval":        tftypes.NewValue(tftypes.Number, nil),
		"quota_time_unit":       tftypes.NewValue(tftypes.String, nil),
		"api_proxy_names":       tftypes.NewValue(apiProxyNamesType, []tftypes.Value{}),
		"additional_properties": tftypes.NewValue(additionalPropsType, []tftypes.Value{}),
	})

	createReq := resource.CreateRequest{Plan: tfsdk.Plan{Schema: s, Raw: planRaw}}
	createResp := &resource.CreateResponse{State: newTestState(t, s)}
	r.Create(ctx, createReq, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create() produced diagnostics: %v", createResp.Diagnostics)
	}

	var got apiProductModel
	createResp.Diagnostics.Append(createResp.State.Get(ctx, &got)...)
	if got.ID.ValueString() != "SampleProduct" {
		t.Errorf("id = %q, want SampleProduct", got.ID.ValueString())
	}

	readReq := resource.ReadRequest{State: createResp.State}
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(ctx, readReq, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", readResp.Diagnostics)
	}

	deleteReq := resource.DeleteRequest{State: readResp.State}
	deleteResp := &resource.DeleteResponse{}
	r.Delete(ctx, deleteReq, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete() produced diagnostics: %v", deleteResp.Diagnostics)
	}
}

package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// SAP answers every update of an API product with 405 (tenant test,
// September 2026), so every attribute a user can set must force a new
// product, and at least one proxy must be linked.
func TestAPIProductResource_SchemaIsReplaceOnly(t *testing.T) {
	var resp resource.SchemaResponse
	NewAPIProductResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	for name, attr := range resp.Schema.Attributes {
		if name == "id" {
			continue
		}
		var modifiers int
		switch a := attr.(type) {
		case schema.StringAttribute:
			modifiers = len(a.PlanModifiers)
		case schema.BoolAttribute:
			modifiers = len(a.PlanModifiers)
		case schema.Int64Attribute:
			modifiers = len(a.PlanModifiers)
		case schema.ListAttribute:
			modifiers = len(a.PlanModifiers)
		case schema.SetNestedAttribute:
			modifiers = len(a.PlanModifiers)
		default:
			t.Fatalf("%s: unexpected attribute type %T", name, attr)
		}
		if modifiers == 0 {
			t.Errorf("%s has no plan modifier; it must force replacement", name)
		}
	}

	proxies := resp.Schema.Attributes["api_proxy_names"].(schema.ListAttribute)
	if !proxies.Required || len(proxies.Validators) == 0 {
		t.Error("api_proxy_names must be required with at least one entry")
	}
}

func TestAPIProductResource_CreateReadDelete(t *testing.T) {
	// The body a tenant returned: navigation properties are __deferred,
	// and SAP filled in version "1" and isPublished true by itself.
	const product = `{"d": {"name": "SampleProduct", "version": "1", "title": "SampleProduct", "description": null, "scope": "", "status_code": "PUBLISHED", "isPublished": true, "isRestricted": false, "quotaCount": null, "quotaInterval": null, "quotaTimeUnit": null, "apiProxies": {"__deferred": {"uri": "x"}}, "additionalProperties": {"__deferred": {"uri": "x"}}}}`
	var postBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &postBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(product))
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			switch {
			case strings.HasSuffix(r.URL.Path, "/apiProxies"):
				_, _ = w.Write([]byte(`{"d": {"results": [{"name": "SampleAPI", "status_code": "REGISTERED"}]}}`))
			case strings.HasSuffix(r.URL.Path, "/additionalProperties"):
				_, _ = w.Write([]byte(`{"d": {"results": []}}`))
			default:
				_, _ = w.Write([]byte(product))
			}
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s: SAP does not support updating a product", r.Method)
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
		"id":              tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":            tftypes.NewValue(tftypes.String, "SampleProduct"),
		"version":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"title":           tftypes.NewValue(tftypes.String, "SampleProduct"),
		"description":     tftypes.NewValue(tftypes.String, nil),
		"scope":           tftypes.NewValue(tftypes.String, nil),
		"status_code":     tftypes.NewValue(tftypes.String, "PUBLISHED"),
		"is_published":    tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
		"is_restricted":   tftypes.NewValue(tftypes.Bool, false),
		"quota_count":     tftypes.NewValue(tftypes.Number, nil),
		"quota_interval":  tftypes.NewValue(tftypes.Number, nil),
		"quota_time_unit": tftypes.NewValue(tftypes.String, nil),
		"api_proxy_names": tftypes.NewValue(apiProxyNamesType, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "SampleAPI"),
		}),
		"additional_properties": tftypes.NewValue(additionalPropsType, nil),
	})

	createReq := resource.CreateRequest{Plan: tfsdk.Plan{Schema: s, Raw: planRaw}}
	createResp := &resource.CreateResponse{State: newTestState(t, s)}
	r.Create(ctx, createReq, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create() produced diagnostics: %v", createResp.Diagnostics)
	}
	if postBody["status_code"] != "PUBLISHED" {
		t.Errorf("status_code sent = %v, want PUBLISHED", postBody["status_code"])
	}
	if proxies, _ := postBody["apiProxies"].([]interface{}); len(proxies) != 1 {
		t.Errorf("apiProxies sent = %v, want one reference", postBody["apiProxies"])
	}

	var got apiProductModel
	createResp.Diagnostics.Append(createResp.State.Get(ctx, &got)...)
	if got.ID.ValueString() != "SampleProduct" {
		t.Errorf("id = %q, want SampleProduct", got.ID.ValueString())
	}
	if got.Version.ValueString() != "1" || !got.IsPublished.ValueBool() {
		t.Errorf("version = %s, is_published = %s; want SAP's values 1 and true", got.Version, got.IsPublished)
	}

	readReq := resource.ReadRequest{State: createResp.State}
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(ctx, readReq, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", readResp.Diagnostics)
	}
	var read apiProductModel
	readResp.Diagnostics.Append(readResp.State.Get(ctx, &read)...)
	if len(read.APIProxyNames) != 1 || read.APIProxyNames[0] != "SampleAPI" {
		t.Errorf("api_proxy_names after read = %v, want [SampleAPI] from /apiProxies", read.APIProxyNames)
	}
	if read.AdditionalProperties != nil {
		t.Errorf("additional_properties after read = %v, want null for SAP's empty list", read.AdditionalProperties)
	}

	deleteReq := resource.DeleteRequest{State: readResp.State}
	deleteResp := &resource.DeleteResponse{}
	r.Delete(ctx, deleteReq, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete() produced diagnostics: %v", deleteResp.Diagnostics)
	}
}

// SAP may list the linked proxies in another order than the configuration;
// that must not show up as a change, because every change replaces the
// product.
func TestKeepListOrder(t *testing.T) {
	cases := []struct {
		prior, current, want []string
	}{
		{[]string{"a", "b"}, []string{"b", "a"}, []string{"a", "b"}},
		{[]string{"a", "b"}, []string{"a", "c"}, []string{"a", "c"}},
		{[]string{"a"}, []string{"a", "b"}, []string{"a", "b"}},
		{nil, []string{"a"}, []string{"a"}},
		{[]string{"a", "a"}, []string{"a", "b"}, []string{"a", "b"}},
	}
	for _, c := range cases {
		if got := keepListOrder(c.prior, c.current); !reflect.DeepEqual(got, c.want) {
			t.Errorf("keepListOrder(%v, %v) = %v, want %v", c.prior, c.current, got, c.want)
		}
	}
}

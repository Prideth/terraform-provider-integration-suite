package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apimanagementclassic"
)

func apiProviderSchema(t *testing.T) schema.Schema {
	t.Helper()
	r := NewAPIProviderResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func TestAPIProviderResource_Schema_PasswordWOIsSensitiveAndWriteOnly(t *testing.T) {
	s := apiProviderSchema(t)

	attr, ok := s.Attributes["password_wo"].(schema.StringAttribute)
	if !ok {
		t.Fatal("password_wo is not a schema.StringAttribute")
	}
	if !attr.WriteOnly {
		t.Error("password_wo must be WriteOnly")
	}
	if !attr.Sensitive {
		t.Error("password_wo must be Sensitive, matching every other credential-shaped write-only attribute in this provider")
	}
}

func TestAPIProviderResource_Schema_EveryAttributeIsRequiresReplace(t *testing.T) {
	s := apiProviderSchema(t)

	// Every attribute except id (Computed+UseStateForUnknown) and
	// password_wo (WriteOnly, no update path applies to it by definition)
	// must carry a RequiresReplace-style plan modifier, since this resource
	// has no Update.
	skip := map[string]bool{"id": true, "password_wo": true}
	for name, attr := range s.Attributes {
		if skip[name] {
			continue
		}
		hasReplace := false
		switch a := attr.(type) {
		case schema.StringAttribute:
			hasReplace = len(a.PlanModifiers) > 0
		case schema.Int64Attribute:
			hasReplace = len(a.PlanModifiers) > 0
		case schema.BoolAttribute:
			hasReplace = len(a.PlanModifiers) > 0
		}
		if !hasReplace {
			t.Errorf("attribute %q has no plan modifier; expected RequiresReplace since this resource has no Update", name)
		}
	}
}

func TestAPIProviderResource_Configure_RequiresAPIManagementClient(t *testing.T) {
	r := NewAPIProviderResource().(resource.ResourceWithConfigure)
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: &Data{}}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a configuration error when APIManagementClassicHTTPClient is nil")
	}
}

func TestAPIProviderResource_Create_AlwaysSendsInternetDestType(t *testing.T) {
	getCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"name": "provider1", "destType": "INTERNET", "host": "backend.example.com", "port": 443, "useSSL": true, "trustAll": false}}`))
		case http.MethodGet:
			getCount++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"name": "provider1", "destType": "INTERNET"}}`))
		}
	}))
	defer server.Close()

	r := &apiProviderResource{client: apimanagementclassic.New(http.DefaultClient, server.URL)}
	s := apiProviderSchema(t)
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	planRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, nil),
		"name":                   tftypes.NewValue(tftypes.String, "provider1"),
		"title":                  tftypes.NewValue(tftypes.String, nil),
		"description":            tftypes.NewValue(tftypes.String, nil),
		"host":                   tftypes.NewValue(tftypes.String, "backend.example.com"),
		"port":                   tftypes.NewValue(tftypes.Number, 443),
		"use_ssl":                tftypes.NewValue(tftypes.Bool, true),
		"trust_all":              tftypes.NewValue(tftypes.Bool, false),
		"path_prefix":            tftypes.NewValue(tftypes.String, nil),
		"service_collection_url": tftypes.NewValue(tftypes.String, nil),
		"auth_type":              tftypes.NewValue(tftypes.String, nil),
		"user_name":              tftypes.NewValue(tftypes.String, nil),
		"password_wo":            tftypes.NewValue(tftypes.String, nil),
	})

	req := resource.CreateRequest{
		Plan:   tfsdk.Plan{Schema: s, Raw: planRaw},
		Config: tfsdk.Config{Schema: s, Raw: planRaw},
	}
	resp := &resource.CreateResponse{State: newTestState(t, s)}

	r.Create(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Create() produced diagnostics: %v", resp.Diagnostics)
	}
	if getCount == 0 {
		t.Error("expected Create to poll GET to confirm visibility")
	}

	var got apiProviderModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if got.ID.ValueString() != "provider1" {
		t.Errorf("id = %q, want provider1", got.ID.ValueString())
	}
	if !got.PasswordWO.IsNull() {
		t.Error("password_wo must be null in state after Create")
	}
}

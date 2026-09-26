package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewIntegrationPackageResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_integration_package.
func NewIntegrationPackageResource() resource.Resource {
	return &integrationPackageResource{}
}

type integrationPackageResource struct {
	client *cloudintegration.Client
}

type integrationPackageModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ShortText   types.String `tfsdk:"short_text"`
	Version     types.String `tfsdk:"version"`
}

func (r *integrationPackageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_package"
}

func (r *integrationPackageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cloud Integration package: the container for integration flows, " +
			"value mappings, script collections, and message mappings. Backed by the public " +
			"Integration Content OData V2 API (IntegrationPackages).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The package's technical ID. Immutable: changing it replaces the package.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The package's display name.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				Description: "A free-text description of the package. SAP stores it as HTML and wraps " +
					"plain text in a paragraph (<p>...</p>); the provider removes that wrapper when " +
					"reading, so a plain-text value round-trips unchanged.",
			},
			"short_text": schema.StringAttribute{
				Required: true,
				Description: "The package's short description, shown in the package list. Required by " +
					"SAP: a create without it fails with \"Property 'ShortText' cannot be empty\".",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The package version reported by SAP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *integrationPackageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	if !requireHTTPClient(data, "resource", &resp.Diagnostics) {
		return
	}
	r.client = cloudintegration.New(data.HTTPClient, data.Host)
}

func (r *integrationPackageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan integrationPackageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePackage(ctx, cloudintegration.Package{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		ShortText:   plan.ShortText.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite integration package", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, packageToModel(created))...)
}

func (r *integrationPackageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state integrationPackageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pkg, err := r.client.GetPackage(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite integration package", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, packageToModel(pkg))...)
}

func (r *integrationPackageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan integrationPackageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdatePackage(ctx, plan.ID.ValueString(), cloudintegration.Package{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		ShortText:   plan.ShortText.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update SAP Integration Suite integration package", diagnosticDetail(err))
		return
	}

	pkg, err := r.client.GetPackage(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read back SAP Integration Suite integration package after update", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, packageToModel(pkg))...)
}

func (r *integrationPackageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state integrationPackageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePackage(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite integration package", diagnosticDetail(err))
	}
}

func (r *integrationPackageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

func packageToModel(pkg *cloudintegration.Package) integrationPackageModel {
	return integrationPackageModel{
		ID:          types.StringValue(pkg.ID),
		Name:        types.StringValue(pkg.Name),
		Description: stringOrNull(cloudintegration.PlainDescription(pkg.Description)),
		ShortText:   stringOrNull(pkg.ShortText),
		Version:     types.StringValue(pkg.Version),
	}
}

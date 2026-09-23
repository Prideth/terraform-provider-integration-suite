package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apimanagementclassic"
)

// NewAPIManagementCertificateStoreReferenceResource returns a fresh
// resource.Resource implementation for
// sapintegrationsuite_api_management_certificate_store_reference.
func NewAPIManagementCertificateStoreReferenceResource() resource.Resource {
	return &apiManagementCertificateStoreReferenceResource{}
}

type apiManagementCertificateStoreReferenceResource struct {
	client *apimanagementclassic.Client
}

type apiManagementCertificateStoreReferenceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	CertificateStoreName types.String `tfsdk:"certificate_store_name"`
	StoreType            types.String `tfsdk:"store_type"`
}

func (r *apiManagementCertificateStoreReferenceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_management_certificate_store_reference"
}

func (r *apiManagementCertificateStoreReferenceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Classic API Management certificate store reference " +
			"(CertificateStoreReferences): a named pointer to an already-existing keystore or " +
			"truststore, used so a virtual host's TLS configuration can be repointed at a new store " +
			"(for certificate rotation) without editing the virtual host itself. Backed by the API " +
			"Portal's Management.svc OData API, confirmed with a full Create/Read/Update/Delete " +
			"lifecycle from SAP's own documented worked examples — the best-confirmed Classic API " +
			"Management resource this provider implements.\n\n" +
			"This resource does not create the keystore or truststore itself, or upload any " +
			"certificate material: SAP documents that as a UI-only operation with no accompanying " +
			"REST API. certificate_store_name must reference a store that already exists (created " +
			"through the SAP Integration Suite UI). This is a distinct remote object from Cloud " +
			"Integration's sapintegrationsuite_certificate/sapintegrationsuite_key_pair, which manage " +
			"actual keystore entries under a completely different API and service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Always equal to name — SAP's confirmed OData key for this entity (CertificateStoreReferences('<name>')).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "The certificate store reference's name.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"certificate_store_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the already-existing keystore or truststore this reference points to. Mutable — SAP's documented PUT updates only this field, which is the entire point of using a reference instead of the literal store name.",
			},
			"store_type": schema.StringAttribute{
				Computed:    true,
				Description: "The referenced store's type, as reported by SAP (confirmed value: \"TRUSTSTORE\").",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *apiManagementCertificateStoreReferenceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	if !requireAPIManagementClassicHTTPClient(data, "resource", &resp.Diagnostics) {
		return
	}
	r.client = apimanagementclassic.New(data.APIManagementClassicHTTPClient, data.APIManagementClassicHost)
}

func (r *apiManagementCertificateStoreReferenceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiManagementCertificateStoreReferenceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateCertificateStoreReference(ctx, apimanagementclassic.CertificateStoreReference{
		Name:                 plan.Name.ValueString(),
		CertificateStoreName: plan.CertificateStoreName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Classic API Management certificate store reference", diagnosticDetail(err))
		return
	}

	plan.ID = types.StringValue(created.Name)
	plan.StoreType = stringOrNull(created.StoreType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiManagementCertificateStoreReferenceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiManagementCertificateStoreReferenceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.client.GetCertificateStoreReference(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Classic API Management certificate store reference", diagnosticDetail(err))
		return
	}

	state.Name = types.StringValue(found.Name)
	state.CertificateStoreName = types.StringValue(found.CertificateStoreName)
	state.StoreType = stringOrNull(found.StoreType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiManagementCertificateStoreReferenceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiManagementCertificateStoreReferenceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateCertificateStoreReference(ctx, plan.ID.ValueString(), plan.CertificateStoreName.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to update Classic API Management certificate store reference", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiManagementCertificateStoreReferenceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiManagementCertificateStoreReferenceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCertificateStoreReference(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete Classic API Management certificate store reference", diagnosticDetail(err))
	}
}

func (r *apiManagementCertificateStoreReferenceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

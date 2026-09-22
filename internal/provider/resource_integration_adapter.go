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

// maxIntegrationAdapterContentBytes bounds how large a local *.esa content
// file this provider will read. SAP does not document a maximum upload
// size for this artifact type anywhere this project could find, so this
// is this provider's own generic protective bound (matching the bound
// already used for script collections and value mappings), not a limit
// SAP itself documents or enforces.
const maxIntegrationAdapterContentBytes = 32 * 1024 * 1024 // 32 MiB

// NewIntegrationAdapterResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_integration_adapter.
func NewIntegrationAdapterResource() resource.Resource {
	return &integrationAdapterResource{}
}

type integrationAdapterResource struct {
	client *cloudintegration.Client
}

type integrationAdapterModel struct {
	ID          types.String `tfsdk:"id"`
	PackageID   types.String `tfsdk:"package_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Application types.String `tfsdk:"application"`
	Content     types.String `tfsdk:"content"`
	ContentHash types.String `tfsdk:"content_hash"`
	Version     types.String `tfsdk:"version"`
}

func (r *integrationAdapterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_adapter"
}

func (r *integrationAdapterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom Integration Adapter design-time artifact: a *.esa archive " +
			"built with the SAP Adapter SDK and imported into a Cloud Integration package. Backed " +
			"by the public Integration Content OData V2 API " +
			"(IntegrationAdapterDesigntimeArtifacts). Cloud Foundry environment only — SAP " +
			"documents this artifact type as unavailable in the Neo environment. This is a " +
			"distinct lifecycle from SAP Business Accelerator Hub prebundled adapters (imported " +
			"and auto-deployed from within the integration flow editor) and from tenant capability " +
			"activation; see docs/guides/integration-adapters.md. This provider's evidence base for " +
			"this specific entity is thinner than for the other design-time artifact types it " +
			"manages — SAP's own \"Example Requests\" documentation for this entity shows only " +
			"Deploy and Delete, not Create or Read — so several fields and the conservative " +
			"replace-on-any-change update model below are corroborated by strong analogy to the " +
			"sibling artifact types in this same API rather than independently confirmed; see the " +
			"guide for the full breakdown.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
				Description: "The integration adapter's technical ID. SAP documents this ID as " +
					"unique across the entire tenant, not just the package, and states that " +
					"importing an ID that already exists is rejected as an error — positive " +
					"evidence that there is no update-in-place via re-import. Immutable: changing " +
					"it replaces the resource. Import uses a composite " +
					"\"<package_id>/<id>\" syntax despite id alone being SAP's confirmed key for " +
					"Delete/Deploy — see package_id below for why.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"package_id": schema.StringAttribute{
				Required: true,
				Description: "ID of the integration package this adapter belongs to at design " +
					"time. Whether an existing adapter can be moved to a different package was not " +
					"confirmed, so changing this replaces the resource rather than attempting an " +
					"unverified move. Required at import time (as the first half of the composite " +
					"import ID) because this project could not confirm that a plain GET by id " +
					"reliably returns the owning package as a queryable property; leaving it " +
					"unrecoverable would otherwise force a destructive replace on the very first " +
					"plan after import, since this attribute is RequiresReplace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "The adapter's display name. SAP's UI documentation states this is " +
					"normally auto-filled from the imported *.esa file's own metadata; this " +
					"provider still requires it explicitly so the configuration self-documents " +
					"which adapter it expects to manage. No confirmed metadata-only update path " +
					"exists (SAP's UI separately mentions editing \"via View metadata\", but no " +
					"public API contract for it was found), so changing this replaces the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Optional: true,
				Description: "The adapter's line-of-business classification, as selected from a " +
					"list in SAP's UI (documented examples: Analytics, CRM, ERP, Finance, HCM, " +
					"Marketing). This provider does not validate it against a fixed set of values: " +
					"SAP's documentation does not confirm whether the underlying OData property is " +
					"a closed enum or a free-form string, and this provider does not add a " +
					"validator without that evidence. Changing it replaces the resource, the same " +
					"conservative treatment as every other metadata field here.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application": schema.StringAttribute{
				Optional: true,
				Description: "The application the adapter provides connectivity for, as selected " +
					"from a list in SAP's UI (SAP's own example: \"Slack\"). Same validation and " +
					"replace-on-change treatment as \"type\", for the same reason.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Path to the local *.esa content file, for example " +
					"\"${path.module}/adapters/custom-sftp-extension.esa\". Required to manage the " +
					"adapter's content; left as-is on import until a matching configuration is " +
					"applied, since SAP does not return a local file path for an existing " +
					"design-time artifact. No public API for replacing an existing adapter's " +
					"content in place was confirmed (see \"id\" above), so changing this replaces " +
					"the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_hash": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "SHA-256 hash of the content file, for example " +
					"filesha256(\"${path.module}/adapters/custom-sftp-extension.esa\"). Terraform " +
					"replaces the adapter when this hash changes, for the same reason as \"content\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Computed: true,
				Description: "The version SAP reports for the adapter, which SAP's UI documents " +
					"as extracted from the *.esa file's own metadata rather than assigned by the " +
					"API on each write.",
			},
		},
	}
}

func (r *integrationAdapterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *integrationAdapterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan integrationAdapterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content, err := readBoundedFile(plan.Content.ValueString(), maxIntegrationAdapterContentBytes)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read integration adapter content file", err.Error())
		return
	}
	if err := verifyContentHash(content, plan.ContentHash.ValueString()); err != nil {
		resp.Diagnostics.AddError("Integration adapter content hash mismatch", err.Error())
		return
	}

	adapter, err := r.client.CreateIntegrationAdapter(ctx, cloudintegration.IntegrationAdapter{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		PackageID:   plan.PackageID.ValueString(),
		Type:        plan.Type.ValueString(),
		Application: plan.Application.ValueString(),
	}, content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite integration adapter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, integrationAdapterToModel(plan.PackageID.ValueString(), adapter, plan))...)
}

func (r *integrationAdapterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state integrationAdapterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adapter, err := r.client.GetIntegrationAdapter(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite integration adapter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, integrationAdapterToModel(state.PackageID.ValueString(), adapter, state))...)
}

// Update is unreachable in practice: every attribute is RequiresReplace,
// since no confirmed public API exists for updating an existing
// integration adapter's metadata or content in place — see the schema
// Description for "id" and docs/guides/integration-adapters.md.
func (r *integrationAdapterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"sapintegrationsuite_integration_adapter does not support in-place updates; "+
			"Terraform should have replaced this resource instead of updating it.",
	)
}

func (r *integrationAdapterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state integrationAdapterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteIntegrationAdapter(ctx, state.ID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite integration adapter", diagnosticDetail(err))
	}
}

// ImportState uses a composite "<package_id>/<id>" import ID, even though
// the confirmed Delete/Deploy operations address an adapter by Id alone
// (see internal/client/cloudintegration/integration_adapter.go). This is a
// deliberate Terraform-side choice, not a claim about SAP's OData key:
// package_id is Required and RequiresReplace in this schema, and this
// project could not confirm that a plain GET by Id reliably returns the
// owning package as a queryable property. Importing by Id alone would
// leave package_id unrecoverable, and since it is RequiresReplace, the
// very first plan after import would want to destroy and recreate the
// adapter purely because Terraform never learned its package — a far
// worse outcome than asking the practitioner to supply a value they
// already know when importing.
func (r *integrationAdapterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	packageID, adapterID, err := splitCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("package_id"), packageID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRootID(), adapterID)...)
}

func integrationAdapterToModel(packageID string, adapter *cloudintegration.IntegrationAdapter, previous integrationAdapterModel) integrationAdapterModel {
	return integrationAdapterModel{
		ID:          types.StringValue(adapter.ID),
		PackageID:   types.StringValue(packageID),
		Name:        types.StringValue(adapter.Name),
		Type:        stringOrNull(adapter.Type),
		Application: stringOrNull(adapter.Application),
		Content:     previous.Content,
		ContentHash: previous.ContentHash,
		Version:     stringOrNull(adapter.Version),
	}
}

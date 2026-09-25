package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// maxPartnerBinaryParameterContentBytes bounds how large a local content
// file this provider will read for a binary parameter, for the same
// reasons documented on maxIntegrationFlowContentBytes. It is deliberately
// larger than partnerdirectory.MaxBinaryParameterValueBytes: this bounds a
// local file read (a basic safety limit against an operator mistake), while
// the SAP-documented 260 KB limit is checked separately, after base64
// encoding, against the actual value SAP will receive.
const maxPartnerBinaryParameterContentBytes = 8 * 1024 * 1024 // 8 MiB

// NewPartnerBinaryParameterResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_partner_binary_parameter.
func NewPartnerBinaryParameterResource() resource.Resource {
	return &partnerBinaryParameterResource{}
}

type partnerBinaryParameterResource struct {
	client *partnerdirectory.Client
}

type partnerBinaryParameterModel struct {
	ID                types.String `tfsdk:"id"`
	PartnerID         types.String `tfsdk:"partner_id"`
	ParameterID       types.String `tfsdk:"parameter_id"`
	ContentType       types.String `tfsdk:"content_type"`
	Content           types.String `tfsdk:"content"`
	ContentHash       types.String `tfsdk:"content_hash"`
	RuntimeLocationID types.String `tfsdk:"runtime_location_id"`
}

func (r *partnerBinaryParameterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_binary_parameter"
}

func (r *partnerBinaryParameterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single Partner Directory binary parameter, a named binary value " +
			"(for example an XSD schema or a certificate) scoped to a partner ID (Pid), uploaded " +
			"from a local content file. Backed by the public Partner Directory OData V2 API " +
			"(BinaryParameters). Partner Directory data is stored unencrypted: do not put " +
			"passwords, secrets, private keys, tokens, or other sensitive information in a binary " +
			"parameter's content — see docs/guides/partner-directory.md. SAP documents a maximum " +
			"decoded value size of 260 KB; larger XML/XSL/XSD content should be stored zipped " +
			"instead (content_type \"zip\" is automatically unzipped by the XML Validator and XSLT " +
			"Mapping steps).",
		Attributes: map[string]schema.Attribute{
			"runtime_location_id": runtimeLocationResourceAttribute(),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<partner_id>/<parameter_id>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"partner_id": schema.StringAttribute{
				Required: true,
				Description: "The Partner ID (Pid) this parameter is scoped to. Does not need to " +
					"already exist: SAP creates it implicitly on the first entity that references it.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parameter_id": schema.StringAttribute{
				Required:    true,
				Description: "The binary parameter's technical ID, unique within its partner.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_type": schema.StringAttribute{
				Required: true,
				Description: "The content's MIME-like type. SAP documents values including " +
					"xml, xsl, xsd, json, text, zip, gz, zlib, and crt, and also accepts " +
					"encoding-suffixed values such as \"xml;encoding=UTF-8\"; this attribute is not " +
					"restricted to a fixed list so a value SAP documents but this provider does not " +
					"yet know about is never rejected client-side.",
			},
			"content": schema.StringAttribute{
				Required: true,
				Description: "Path to the local content file, for example " +
					"\"${path.module}/partner-directory/order-schema.xsd\".",
			},
			"content_hash": schema.StringAttribute{
				Required: true,
				Description: "SHA-256 hash of the content file, for example " +
					"filesha256(\"${path.module}/partner-directory/order-schema.xsd\"). Terraform " +
					"only re-uploads the file when this hash changes.",
			},
		},
	}
}

func (r *partnerBinaryParameterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = partnerdirectory.New(data.HTTPClient, data.Host)
}

// readAndValidateBinaryParameterContent reads the local content file and
// checks it before any network call: first the content hash (a
// configuration mistake caught early), then SAP's documented 260 KB limit
// on the raw (pre-base64) value, so an obviously oversized payload never
// reaches the API at all.
func readAndValidateBinaryParameterContent(path, expectedHash string) ([]byte, error) {
	content, err := readBoundedFile(path, maxPartnerBinaryParameterContentBytes)
	if err != nil {
		return nil, err
	}
	if err := verifyContentHash(content, expectedHash); err != nil {
		return nil, err
	}
	if len(content) > partnerdirectory.MaxBinaryParameterValueBytes {
		return nil, fmt.Errorf(
			"content is %d bytes, which exceeds SAP's documented 260 KB (%d byte) limit for a Partner Directory binary parameter value; store larger XML/XSL/XSD content zipped instead",
			len(content), partnerdirectory.MaxBinaryParameterValueBytes,
		)
	}
	return content, nil
}

func (r *partnerBinaryParameterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan partnerBinaryParameterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, plan.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	content, err := readAndValidateBinaryParameterContent(plan.Content.ValueString(), plan.ContentHash.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Partner Directory binary parameter content file", err.Error())
		return
	}

	created, err := client.CreateBinaryParameter(ctx, plan.PartnerID.ValueString(), plan.ParameterID.ValueString(), plan.ContentType.ValueString(), content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite Partner Directory binary parameter", diagnosticDetail(err))
		return
	}

	m := binaryParameterToModel(created, plan)
	m.RuntimeLocationID = plan.RuntimeLocationID
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *partnerBinaryParameterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state partnerBinaryParameterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, state.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	bp, err := client.GetBinaryParameter(ctx, state.PartnerID.ValueString(), state.ParameterID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory binary parameter", diagnosticDetail(err))
		return
	}

	m := binaryParameterToModel(bp, state)
	m.RuntimeLocationID = state.RuntimeLocationID
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *partnerBinaryParameterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan partnerBinaryParameterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, plan.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	content, err := readAndValidateBinaryParameterContent(plan.Content.ValueString(), plan.ContentHash.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Partner Directory binary parameter content file", err.Error())
		return
	}

	err = client.UpdateBinaryParameter(ctx, plan.PartnerID.ValueString(), plan.ParameterID.ValueString(), plan.ContentType.ValueString(), content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update SAP Integration Suite Partner Directory binary parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerBinaryParameterModel{
		ID:          types.StringValue(plan.PartnerID.ValueString() + "/" + plan.ParameterID.ValueString()),
		PartnerID:   plan.PartnerID,
		ParameterID: plan.ParameterID,
		ContentType: plan.ContentType,
		Content:     plan.Content,
		ContentHash: plan.ContentHash,
	})...)
}

func (r *partnerBinaryParameterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerBinaryParameterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, state.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	err := client.DeleteBinaryParameter(ctx, state.PartnerID.ValueString(), state.ParameterID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite Partner Directory binary parameter", diagnosticDetail(err))
	}
}

func (r *partnerBinaryParameterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	loc, parts, err := splitLocatedImportID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	partnerID, parameterID := parts[0], parts[1]
	setImportedRuntimeLocation(ctx, loc, resp.State.SetAttribute, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("partner_id"), partnerID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("parameter_id"), parameterID)...)
}

// binaryParameterToModel builds the resource model from the API response,
// preserving the local content/content_hash from previous (Terraform's own
// state or plan), the same pattern used by every other file-based resource
// in this provider: SAP does not return a local file path, so there is
// nothing to read it back from.
func binaryParameterToModel(bp *partnerdirectory.BinaryParameter, previous partnerBinaryParameterModel) partnerBinaryParameterModel {
	return partnerBinaryParameterModel{
		ID:          types.StringValue(bp.Pid + "/" + bp.Id),
		PartnerID:   types.StringValue(bp.Pid),
		ParameterID: types.StringValue(bp.Id),
		ContentType: types.StringValue(bp.ContentType),
		Content:     previous.Content,
		ContentHash: previous.ContentHash,
	}
}

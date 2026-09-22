package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewAlternativePartnerResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_alternative_partner.
func NewAlternativePartnerResource() resource.Resource {
	return &alternativePartnerResource{}
}

type alternativePartnerResource struct {
	client *partnerdirectory.Client
}

type alternativePartnerModel struct {
	ID         types.String `tfsdk:"id"`
	Agency     types.String `tfsdk:"agency"`
	Scheme     types.String `tfsdk:"scheme"`
	ExternalID types.String `tfsdk:"external_id"`
	PartnerID  types.String `tfsdk:"partner_id"`
}

func (r *alternativePartnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alternative_partner"
}

func (r *alternativePartnerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Partner Directory alternative partner mapping: an external " +
			"identity tuple (agency, scheme, external_id) resolved to an internal Partner ID. " +
			"Backed by the public Partner Directory OData V2 API (AlternativePartners). SAP's " +
			"actual entity key is a hex encoding of agency/scheme/external_id " +
			"(Hexagency/Hexscheme/Hexid); this provider computes that internally and never " +
			"exposes it as something a practitioner sets directly.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Composite identifier in the form " +
					"\"<hex_agency>/<hex_scheme>/<hex_external_id>\" (each segment hex-encoded so " +
					"the identifier is well-formed regardless of what characters agency, scheme, " +
					"and external_id contain).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agency": schema.StringAttribute{
				Required:    true,
				Description: "The external scheme agency identifier, for example \"Sender_1\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scheme": schema.StringAttribute{
				Required:    true,
				Description: "The external identification scheme, for example \"SenderInterface\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"external_id": schema.StringAttribute{
				Required:    true,
				Description: "The external identifier within agency/scheme, for example \"Interface_1\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"partner_id": schema.StringAttribute{
				Required: true,
				Description: "The internal Partner ID (Pid) this external identity resolves to. " +
					"Mutable in place: SAP documents PUT for repointing an existing mapping at a " +
					"different Pid.",
			},
		},
	}
}

func (r *alternativePartnerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *alternativePartnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alternativePartnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateAlternativePartner(ctx, partnerdirectory.AlternativePartner{
		Agency: plan.Agency.ValueString(),
		Scheme: plan.Scheme.ValueString(),
		Id:     plan.ExternalID.ValueString(),
		Pid:    plan.PartnerID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite alternative partner", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, alternativePartnerToModel(created))...)
}

func (r *alternativePartnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alternativePartnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ap, err := r.client.GetAlternativePartner(ctx, state.Agency.ValueString(), state.Scheme.ValueString(), state.ExternalID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite alternative partner", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, alternativePartnerToModel(ap))...)
}

func (r *alternativePartnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alternativePartnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateAlternativePartner(ctx, plan.Agency.ValueString(), plan.Scheme.ValueString(), plan.ExternalID.ValueString(), plan.PartnerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update SAP Integration Suite alternative partner", diagnosticDetail(err))
		return
	}

	id := alternativePartnerID(plan.Agency.ValueString(), plan.Scheme.ValueString(), plan.ExternalID.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, alternativePartnerModel{
		ID:         types.StringValue(id),
		Agency:     plan.Agency,
		Scheme:     plan.Scheme,
		ExternalID: plan.ExternalID,
		PartnerID:  plan.PartnerID,
	})...)
}

func (r *alternativePartnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alternativePartnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAlternativePartner(ctx, state.Agency.ValueString(), state.Scheme.ValueString(), state.ExternalID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite alternative partner", diagnosticDetail(err))
	}
}

// ImportState accepts "<hex_agency>/<hex_scheme>/<hex_external_id>", the
// same three hex-encoded segments alternativePartnerID builds, so an
// import ID is always well-formed no matter what characters the plain
// agency/scheme/external_id strings contain.
func (r *alternativePartnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	agency, scheme, externalID, err := parseAlternativePartnerImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("agency"), agency)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("scheme"), scheme)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("external_id"), externalID)...)
}

// alternativePartnerID builds this resource's Terraform ID: the three hex
// key components joined by "/". Hex-encoded segments can never themselves
// contain "/", so this is unambiguous to split back apart regardless of
// what the plain agency/scheme/external_id values are.
func alternativePartnerID(agency, scheme, externalID string) string {
	hexAgency, hexScheme, hexID := partnerdirectory.EncodeAlternativePartnerKey(agency, scheme, externalID)
	return hexAgency + "/" + hexScheme + "/" + hexID
}

// parseAlternativePartnerImportID splits a "<hex_agency>/<hex_scheme>/<hex_id>"
// import ID and decodes each segment back to its plain string.
func parseAlternativePartnerImportID(id string) (agency, scheme, externalID string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf(
			"expected an import ID in the form \"<hex_agency>/<hex_scheme>/<hex_external_id>\" (three hex-encoded segments), got %q",
			id,
		)
	}

	agency, err = partnerdirectory.DecodeAlternativePartnerKeyComponent(parts[0])
	if err != nil {
		return "", "", "", err
	}
	scheme, err = partnerdirectory.DecodeAlternativePartnerKeyComponent(parts[1])
	if err != nil {
		return "", "", "", err
	}
	externalID, err = partnerdirectory.DecodeAlternativePartnerKeyComponent(parts[2])
	if err != nil {
		return "", "", "", err
	}
	return agency, scheme, externalID, nil
}

func alternativePartnerToModel(ap *partnerdirectory.AlternativePartner) alternativePartnerModel {
	return alternativePartnerModel{
		ID:         types.StringValue(alternativePartnerID(ap.Agency, ap.Scheme, ap.Id)),
		Agency:     types.StringValue(ap.Agency),
		Scheme:     types.StringValue(ap.Scheme),
		ExternalID: types.StringValue(ap.Id),
		PartnerID:  types.StringValue(ap.Pid),
	}
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/securitycontent"
)

// NewOAuth2ClientCredentialDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_oauth2_client_credential.
func NewOAuth2ClientCredentialDataSource() datasource.DataSource {
	return &oauth2ClientCredentialDataSource{}
}

type oauth2ClientCredentialDataSource struct {
	client *securitycontent.Client
}

type oauth2ClientCredentialDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Description          types.String `tfsdk:"description"`
	TokenServiceURL      types.String `tfsdk:"token_service_url"`
	ClientID             types.String `tfsdk:"client_id"`
	Scope                types.String `tfsdk:"scope"`
	ClientAuthentication types.String `tfsdk:"client_authentication"`
	ScopeContentType     types.String `tfsdk:"scope_content_type"`
	Resource             types.String `tfsdk:"resource"`
	Audience             types.String `tfsdk:"audience"`
}

func (d *oauth2ClientCredentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth2_client_credential"
}

func (d *oauth2ClientCredentialDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Security Content \"OAuth2 Client Credentials\" artifact by " +
			"its name. Never returns the client secret: SAP's Security Content API does not document " +
			"returning a stored credential's secret, and this data source has no field for one even " +
			"if it did.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The credential artifact's name (its OData key and adapter alias).",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The credential artifact's free-text description.",
			},
			"token_service_url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of the OAuth2 authorization server that issues the access token.",
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OAuth2 client ID registered with the token service.",
			},
			"scope": schema.StringAttribute{
				Computed:    true,
				Description: "OAuth2 scope requested, if the token service requires one.",
			},
			"client_authentication": schema.StringAttribute{
				Computed:    true,
				Description: "How the client ID and secret are sent to the token service, as SAP stores it.",
			},
			"scope_content_type": schema.StringAttribute{
				Computed:    true,
				Description: "Content type of the token request, as SAP stores it.",
			},
			"resource": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier sent to the token service, if any.",
			},
			"audience": schema.StringAttribute{
				Computed:    true,
				Description: "Audience identifier sent to the token service, if any.",
			},
		},
	}
}

func (d *oauth2ClientCredentialDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	if !requireHTTPClient(data, "data source", &resp.Diagnostics) {
		return
	}
	d.client = securitycontent.New(data.HTTPClient, data.Host)
}

func (d *oauth2ClientCredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config oauth2ClientCredentialDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cred, err := d.client.GetOAuth2ClientCredential(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite OAuth2 client credential", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, oauth2ClientCredentialDataSourceModel{
		ID:                   types.StringValue(cred.Name),
		Description:          stringOrNull(cred.Description),
		TokenServiceURL:      types.StringValue(cred.TokenServiceURL),
		ClientID:             types.StringValue(cred.ClientID),
		Scope:                stringOrNull(cred.Scope),
		ClientAuthentication: stringOrNull(cred.ClientAuthentication),
		ScopeContentType:     stringOrNull(cred.ScopeContentType),
		Resource:             stringOrNull(cred.Resource),
		Audience:             stringOrNull(cred.Audience),
	})...)
}

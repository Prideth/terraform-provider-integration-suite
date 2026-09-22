// Package provider implements the sapintegrationsuite Terraform provider.
package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/auth"
	sapthttp "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/http"
)

// Version is set at build time via -ldflags by GoReleaser; it defaults to
// "dev" for local builds.
var Version = "dev"

// New returns a provider.Provider factory for the terraform-plugin-framework
// server, capturing the build version for the User-Agent header.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &sapIntegrationSuiteProvider{version: version}
	}
}

type sapIntegrationSuiteProvider struct {
	version string
}

// providerModel mirrors the provider block's schema.
type providerModel struct {
	Host  types.String `tfsdk:"host"`
	OAuth *oauthModel  `tfsdk:"oauth"`
}

type oauthModel struct {
	TokenURL     types.String `tfsdk:"token_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// Data is the fully resolved provider configuration made available to every
// resource and data source via Configure. Capability-specific clients (Cloud
// Integration, API Management, ...) are built lazily by each resource from
// these shared ingredients, rather than eagerly during provider Configure,
// so that a tenant missing one capability does not block using another.
type Data struct {
	Host       string
	HTTPClient *sapthttp.Client
	Version    string
}

func (p *sapIntegrationSuiteProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sapintegrationsuite"
	resp.Version = p.version
}

func (p *sapIntegrationSuiteProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configures the SAP Integration Suite provider. This provider administers " +
			"content and capabilities inside an already-provisioned SAP Integration Suite tenant; " +
			"it does not create BTP subaccounts, entitlements, or the Integration Suite subscription " +
			"itself. Use the official SAP BTP provider for those.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the SAP Integration Suite tenant used for Cloud Integration " +
					"APIs, for example https://<tenant>.it-cpi<...>.cfapps.<region>.hana.ondemand.com. " +
					"Can also be set via the SAP_INTEGRATION_SUITE_HOST environment variable.",
			},
		},
		Blocks: map[string]schema.Block{
			"oauth": schema.SingleNestedBlock{
				Description: "OAuth 2.0 client credentials used to authenticate against the SAP " +
					"Integration Suite APIs.",
				Attributes: map[string]schema.Attribute{
					"token_url": schema.StringAttribute{
						Optional: true,
						Description: "OAuth 2.0 token endpoint URL. Can also be set via the " +
							"SAP_INTEGRATION_SUITE_TOKEN_URL environment variable.",
					},
					"client_id": schema.StringAttribute{
						Optional:    true,
						Description: "OAuth 2.0 client ID. Can also be set via the SAP_INTEGRATION_SUITE_CLIENT_ID environment variable.",
					},
					"client_secret": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "OAuth 2.0 client secret. Can also be set via the SAP_INTEGRATION_SUITE_CLIENT_SECRET environment variable.",
					},
				},
			},
		},
	}
}

func (p *sapIntegrationSuiteProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := stringOrEnv(config.Host, "SAP_INTEGRATION_SUITE_HOST")
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing SAP Integration Suite host",
			"Set the host attribute or the SAP_INTEGRATION_SUITE_HOST environment variable.",
		)
	}

	var tokenURL, clientID, clientSecret types.String
	if config.OAuth != nil {
		tokenURL = config.OAuth.TokenURL
		clientID = config.OAuth.ClientID
		clientSecret = config.OAuth.ClientSecret
	}

	oauthCfg := auth.Config{
		TokenURL:     stringOrEnv(tokenURL, "SAP_INTEGRATION_SUITE_TOKEN_URL"),
		ClientID:     stringOrEnv(clientID, "SAP_INTEGRATION_SUITE_CLIENT_ID"),
		ClientSecret: stringOrEnv(clientSecret, "SAP_INTEGRATION_SUITE_CLIENT_SECRET"),
	}

	if oauthCfg.TokenURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oauth").AtName("token_url"),
			"Missing OAuth token URL",
			"Set oauth.token_url or the SAP_INTEGRATION_SUITE_TOKEN_URL environment variable.",
		)
	}
	if oauthCfg.ClientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oauth").AtName("client_id"),
			"Missing OAuth client ID",
			"Set oauth.client_id or the SAP_INTEGRATION_SUITE_CLIENT_ID environment variable.",
		)
	}
	if oauthCfg.ClientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oauth").AtName("client_secret"),
			"Missing OAuth client secret",
			"Set oauth.client_secret or the SAP_INTEGRATION_SUITE_CLIENT_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	authenticatedClient, invalidateToken, err := oauthCfg.HTTPClient(ctx, http.DefaultClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to configure SAP Integration Suite authentication", err.Error())
		return
	}

	data := &Data{
		Host: host,
		HTTPClient: sapthttp.New(sapthttp.Config{
			Transport:       authenticatedClient,
			UserAgent:       sapthttp.UserAgent(p.version),
			InvalidateToken: invalidateToken,
		}),
		Version: p.version,
	}

	resp.DataSourceData = data
	resp.ResourceData = data
}

func (p *sapIntegrationSuiteProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewIntegrationPackageResource,
		NewIntegrationFlowResource,
		NewIntegrationFlowDeploymentResource,
		NewAccessPolicyResource,
		NewAccessPolicyReferenceResource,
	}
}

func (p *sapIntegrationSuiteProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewIntegrationPackageDataSource,
	}
}

// stringOrEnv returns value's string contents if it is known and non-empty,
// otherwise falls back to the named environment variable.
func stringOrEnv(value types.String, envVar string) string {
	if !value.IsNull() && !value.IsUnknown() && value.ValueString() != "" {
		return value.ValueString()
	}
	return os.Getenv(envVar)
}

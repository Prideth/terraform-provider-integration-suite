package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewCustomTagConfigurationDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_custom_tag_configuration.
func NewCustomTagConfigurationDataSource() datasource.DataSource {
	return &customTagConfigurationDataSource{}
}

type customTagConfigurationDataSource struct {
	client *cloudintegration.Client
}

type customTagConfigurationDataSourceModel struct {
	ID   types.String     `tfsdk:"id"`
	Tags []customTagModel `tfsdk:"tags"`
}

func (d *customTagConfigurationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_tag_configuration"
}

func (d *customTagConfigurationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the tenant's current Custom Tag Configuration — the tenant-wide set " +
			"of custom tags integration package owners are asked, or required, to classify their " +
			"packages with. Takes no arguments: SAP documents exactly one configuration per " +
			"tenant. Useful for discovering an existing configuration before adopting it with " +
			"sapintegrationsuite_custom_tag_configuration, or for reading it without taking " +
			"ownership at all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always \"CustomTags\" — SAP's confirmed fixed identifier for this configuration.",
			},
			"tags": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Every custom tag currently defined on the tenant.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The tag's name.",
						},
						"mandatory": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether an integration package owner must set this tag.",
						},
						"permitted_values": schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "The closed list of values this tag accepts, if SAP reports one.",
						},
					},
				},
			},
		},
	}
}

func (d *customTagConfigurationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = cloudintegration.New(data.HTTPClient, data.Host)
}

func (d *customTagConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tags, err := d.client.GetCustomTagConfiguration(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read the SAP Integration Suite custom tag configuration", diagnosticDetail(err))
		return
	}

	// customTagConfigurationDataSourceModel and customTagConfigurationModel
	// (defined in resource_custom_tag_configuration.go) have identical
	// fields, so this is a plain type conversion, not a literal rebuild.
	model := customTagConfigurationDataSourceModel(customTagConfigurationToModel(tags))
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

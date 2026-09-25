package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/securitycontent"
)

// NewKeystoreEntriesDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_keystore_entries.
func NewKeystoreEntriesDataSource() datasource.DataSource {
	return &keystoreEntriesDataSource{}
}

type keystoreEntriesDataSource struct {
	client *securitycontent.Client
}

type keystoreEntriesDataSourceModel struct {
	Entries           []keystoreEntryModel `tfsdk:"entries"`
	RuntimeLocationID types.String         `tfsdk:"runtime_location_id"`
}

func (d *keystoreEntriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keystore_entries"
}

func (d *keystoreEntriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every entry of the tenant keystore, both tenant-owned and SAP-owned; " +
			"owner tells them apart. Useful for brownfield discovery before importing " +
			"sapintegrationsuite_certificate or sapintegrationsuite_key_pair resources, and for " +
			"finding certificates that are about to expire. Entries are sorted by alias, since SAP " +
			"does not document a response order.",
		Attributes: map[string]schema.Attribute{
			"runtime_location_id": runtimeLocationDataSourceAttribute(),
			"entries": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Every entry in the tenant keystore, sorted by alias.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: keystoreEntryAttributes(false),
				},
			},
		},
	}
}

func (d *keystoreEntriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *keystoreEntriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config keystoreEntriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(d.client, config.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	entries, err := client.ListKeystoreEntries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list SAP Integration Suite keystore entries", diagnosticDetail(err))
		return
	}

	models := make([]keystoreEntryModel, 0, len(entries))
	for _, e := range entries {
		entry := e
		models = append(models, keystoreEntryToModel(&entry))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, keystoreEntriesDataSourceModel{Entries: models, RuntimeLocationID: config.RuntimeLocationID})...)
}

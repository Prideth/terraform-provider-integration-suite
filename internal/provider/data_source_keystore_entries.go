package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

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
	Entries []keystoreEntryModel `tfsdk:"entries"`
}

func (d *keystoreEntriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keystore_entries"
}

func (d *keystoreEntriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every entry in the tenant keystore — both tenant-administrator-owned " +
			"and SAP-owned entries; SAP's API exposes no field distinguishing the two, so this data " +
			"source cannot filter by ownership. Useful for brownfield discovery before importing " +
			"individual sapintegrationsuite_certificate or sapintegrationsuite_key_pair resources. " +
			"Backed by the public Security Content OData V2 API (KeystoreEntries); entries are " +
			"sorted by alias, since SAP does not document a guaranteed response order.",
		Attributes: map[string]schema.Attribute{
			"entries": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Every entry in the tenant keystore, sorted by alias.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"alias": schema.StringAttribute{
							Computed:    true,
							Description: "The keystore entry's alias.",
						},
						"hex_alias": schema.StringAttribute{
							Computed:    true,
							Description: "The lowercase hex encoding of alias's UTF-8 bytes — SAP's actual OData key for this entity.",
						},
						"key_type": schema.StringAttribute{
							Computed:    true,
							Description: "The entry's key type, exactly as SAP returns it (for example \"RSA\", \"DSA\", \"EC\").",
						},
						"key_size": schema.Int64Attribute{
							Computed:    true,
							Description: "The entry's key size in bits.",
						},
						"valid_not_before": schema.StringAttribute{
							Computed:    true,
							Description: "The lower boundary of the certificate's validity period, exactly as SAP returns it.",
						},
						"valid_not_after": schema.StringAttribute{
							Computed:    true,
							Description: "The upper boundary of the certificate's validity period, exactly as SAP returns it.",
						},
					},
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

func (d *keystoreEntriesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	entries, err := d.client.ListKeystoreEntries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list SAP Integration Suite keystore entries", diagnosticDetail(err))
		return
	}

	models := make([]keystoreEntryModel, 0, len(entries))
	for _, e := range entries {
		entry := e
		models = append(models, keystoreEntryToModel(&entry))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, keystoreEntriesDataSourceModel{Entries: models})...)
}

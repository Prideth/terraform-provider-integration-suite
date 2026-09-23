package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/securitycontent"
)

// NewKeystoreEntryDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_keystore_entry.
func NewKeystoreEntryDataSource() datasource.DataSource {
	return &keystoreEntryDataSource{}
}

type keystoreEntryDataSource struct {
	client *securitycontent.Client
}

func (d *keystoreEntryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keystore_entry"
}

func (d *keystoreEntryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a single entry (certificate, SAP-generated key pair, or other " +
			"RSA/DSA/EC-keyed entry) from the tenant keystore by its alias. Backed by the public " +
			"Security Content OData V2 API (KeystoreEntries). Read-only: this provider exposes no " +
			"way to tell, from this API alone, whether an entry is owned by the tenant " +
			"administrator or by SAP — see docs/guides/security-content.md for how " +
			"sapintegrationsuite_certificate and sapintegrationsuite_key_pair handle that gap for " +
			"the entries they manage.",
		Attributes: map[string]schema.Attribute{
			"alias": schema.StringAttribute{
				Required:    true,
				Description: "The keystore entry's alias.",
			},
			"hex_alias": schema.StringAttribute{
				Computed: true,
				Description: "The lowercase hex encoding of alias's UTF-8 bytes — SAP's actual " +
					"OData key for this entity. Exposed only as informational metadata; you never " +
					"need to compute or supply it yourself.",
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
	}
}

func (d *keystoreEntryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *keystoreEntryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config keystoreEntryModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry, err := d.client.GetKeystoreEntry(ctx, config.Alias.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite keystore entry", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, keystoreEntryToModel(entry))...)
}

// keystoreEntryModel is shared between the single and collection keystore
// entry data sources.
type keystoreEntryModel struct {
	Alias          types.String `tfsdk:"alias"`
	HexAlias       types.String `tfsdk:"hex_alias"`
	KeyType        types.String `tfsdk:"key_type"`
	KeySize        types.Int64  `tfsdk:"key_size"`
	ValidNotBefore types.String `tfsdk:"valid_not_before"`
	ValidNotAfter  types.String `tfsdk:"valid_not_after"`
}

func keystoreEntryToModel(entry *securitycontent.KeystoreEntry) keystoreEntryModel {
	return keystoreEntryModel{
		Alias:          types.StringValue(entry.Alias),
		HexAlias:       types.StringValue(entry.Hexalias),
		KeyType:        stringOrNull(entry.KeyType),
		KeySize:        int64OrNull(entry.KeySize),
		ValidNotBefore: stringOrNull(entry.ValidNotBefore),
		ValidNotAfter:  stringOrNull(entry.ValidNotAfter),
	}
}

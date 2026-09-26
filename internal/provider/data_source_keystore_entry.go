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
		Description: "Reads one entry (certificate, key pair or SSH key) of the tenant keystore by " +
			"its alias, with the certificate details SAP stores for it: subject and issuer, " +
			"validity, fingerprints, and who owns and last changed the entry. Backed by the " +
			"KeystoreEntries entity of the Security Content OData V2 API. Read-only.",
		Attributes: keystoreEntryDataSourceAttributes(),
	}
}

// keystoreEntryAttributes is shared by the single and collection keystore
// entry data sources. aliasRequired selects the single-entry lookup form.
func keystoreEntryAttributes(aliasRequired bool) map[string]schema.Attribute {
	computed := func(description string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, Description: description}
	}
	return map[string]schema.Attribute{
		"alias": schema.StringAttribute{
			Required:    aliasRequired,
			Computed:    !aliasRequired,
			Description: "The keystore entry's alias.",
		},
		"hex_alias": computed("Lowercase hex encoding of the alias's UTF-8 bytes, which SAP uses as " +
			"the OData key. Informational; you never need to supply it."),
		"entry_type": computed("Kind of entry as SAP reports it in Type, observed on a tenant as " +
			"\"Certificate\" or \"Key Pair\" (with a space)."),
		"owner": computed("Who owns the entry as SAP reports it, \"SAP\" for the root certificates " +
			"and the key pair SAP delivers. SAP-owned entries are managed by SAP and should not be " +
			"changed by tenant automation."),
		"status": computed("Status SAP reports for the entry, for example \"unchanged\" for " +
			"SAP-delivered entries nobody modified."),
		"key_type":            computed("Key type, for example \"RSA\", \"DSA\" or \"EC\"."),
		"key_size":            schema.Int64Attribute{Computed: true, Description: "Key size in bits."},
		"elliptic_curve":      computed("Curve name for EC keys, or null."),
		"signature_algorithm": computed("Signature algorithm of the certificate."),
		"serial_number":       computed("Serial number of the certificate."),
		"subject_dn":          computed("Subject distinguished name of the certificate."),
		"issuer_dn":           computed("Issuer distinguished name of the certificate."),
		"certificate_version": schema.Int64Attribute{Computed: true, Description: "X.509 version of the certificate."},
		"validity":            computed("Validity state SAP derives from the validity period. Observed empty (null) on a tenant for valid entries."),
		"valid_not_before":    computed("Start of the certificate's validity period, RFC 3339 in UTC."),
		"valid_not_after":     computed("End of the certificate's validity period, RFC 3339 in UTC."),
		"fingerprint_sha1":    computed("SHA-1 fingerprint of the certificate as SAP reports it."),
		"fingerprint_sha256":  computed("SHA-256 fingerprint of the certificate as SAP reports it."),
		"fingerprint_sha512":  computed("SHA-512 fingerprint of the certificate as SAP reports it."),
		"created_by":          computed("User who created the entry."),
		"created_time":        computed("When the entry was created, RFC 3339 in UTC."),
		"last_modified_by":    computed("User who last changed the entry."),
		"last_modified_time":  computed("When the entry was last changed, RFC 3339 in UTC."),
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
	var config keystoreEntryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(d.client, config.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	entry, err := client.GetKeystoreEntry(ctx, config.Alias.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite keystore entry", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, keystoreEntryDataSourceModel{
		keystoreEntryModel: keystoreEntryToModel(entry),
		RuntimeLocationID:  config.RuntimeLocationID,
	})...)
}

// keystoreEntryModel is shared between the single and collection keystore
// entry data sources.
type keystoreEntryModel struct {
	Alias              types.String `tfsdk:"alias"`
	HexAlias           types.String `tfsdk:"hex_alias"`
	EntryType          types.String `tfsdk:"entry_type"`
	Owner              types.String `tfsdk:"owner"`
	Status             types.String `tfsdk:"status"`
	KeyType            types.String `tfsdk:"key_type"`
	KeySize            types.Int64  `tfsdk:"key_size"`
	EllipticCurve      types.String `tfsdk:"elliptic_curve"`
	SignatureAlgorithm types.String `tfsdk:"signature_algorithm"`
	SerialNumber       types.String `tfsdk:"serial_number"`
	SubjectDN          types.String `tfsdk:"subject_dn"`
	IssuerDN           types.String `tfsdk:"issuer_dn"`
	CertificateVersion types.Int64  `tfsdk:"certificate_version"`
	Validity           types.String `tfsdk:"validity"`
	ValidNotBefore     types.String `tfsdk:"valid_not_before"`
	ValidNotAfter      types.String `tfsdk:"valid_not_after"`
	FingerprintSha1    types.String `tfsdk:"fingerprint_sha1"`
	FingerprintSha256  types.String `tfsdk:"fingerprint_sha256"`
	FingerprintSha512  types.String `tfsdk:"fingerprint_sha512"`
	CreatedBy          types.String `tfsdk:"created_by"`
	CreatedTime        types.String `tfsdk:"created_time"`
	LastModifiedBy     types.String `tfsdk:"last_modified_by"`
	LastModifiedTime   types.String `tfsdk:"last_modified_time"`
}

func keystoreEntryToModel(entry *securitycontent.KeystoreEntry) keystoreEntryModel {
	return keystoreEntryModel{
		Alias:              types.StringValue(entry.Alias),
		HexAlias:           types.StringValue(entry.Hexalias),
		EntryType:          stringOrNull(entry.Type),
		Owner:              stringOrNull(entry.Owner),
		Status:             stringOrNull(entry.Status),
		KeyType:            stringOrNull(entry.KeyType),
		KeySize:            int64OrNull(entry.KeySize),
		EllipticCurve:      stringOrNull(entry.EllipticCurve),
		SignatureAlgorithm: stringOrNull(entry.SignatureAlgorithm),
		SerialNumber:       stringOrNull(entry.SerialNumber),
		SubjectDN:          stringOrNull(entry.SubjectDN),
		IssuerDN:           stringOrNull(entry.IssuerDN),
		CertificateVersion: int64OrNull(entry.Version),
		Validity:           stringOrNull(entry.Validity),
		ValidNotBefore:     odataDateToRFC3339(entry.ValidNotBefore),
		ValidNotAfter:      odataDateToRFC3339(entry.ValidNotAfter),
		FingerprintSha1:    stringOrNull(entry.FingerprintSha1),
		FingerprintSha256:  stringOrNull(entry.FingerprintSha256),
		FingerprintSha512:  stringOrNull(entry.FingerprintSha512),
		CreatedBy:          stringOrNull(entry.CreatedBy),
		CreatedTime:        odataDateToRFC3339(entry.CreatedTime),
		LastModifiedBy:     stringOrNull(entry.LastModifiedBy),
		LastModifiedTime:   odataDateToRFC3339(entry.LastModifiedTime),
	}
}

// keystoreEntryDataSourceModel adds the lookup-only runtime_location_id to the
// entry fields shared with the collection data source's list items.
type keystoreEntryDataSourceModel struct {
	keystoreEntryModel
	RuntimeLocationID types.String `tfsdk:"runtime_location_id"`
}

func keystoreEntryDataSourceAttributes() map[string]schema.Attribute {
	attrs := keystoreEntryAttributes(true)
	attrs["runtime_location_id"] = runtimeLocationDataSourceAttribute()
	return attrs
}

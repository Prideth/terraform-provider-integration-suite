package securitycontent

import (
	"context"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const (
	certificateResourcesEntitySet = "CertificateResources"

	// certificatePEMContentType is sent as the request Content-Type when
	// importing or updating a certificate. SAP documents it for the export
	// (Export Certificate); a tenant test in September 2026 confirmed that
	// the import accepts it with a plain PEM body.
	certificatePEMContentType = "application/pkix-cert"
)

// Query options of the certificate import, confirmed on a tenant in
// September 2026:
//
//   - Without fingerprintVerified=true, SAP answered a self-signed
//     certificate with 409 and Status "notImported": it parses the
//     certificate and waits for the fingerprint to be confirmed, as the
//     UI's import dialog does. With it, the import answered 204.
//   - Replacing the certificate of an existing alias answered 400 "Entry
//     with alias ... already exists in keystore" unless update=true was
//     sent as well; with it, 204 and the new certificate read back.
//   - returnKeystoreEntries=false keeps the response empty.
//
// SAP's Help page for "Import and Update Certificate" could not be read
// without a browser; the option names come from the tenant test and a
// community blog that uses the same request.
const (
	certificateImportQuery = "fingerprintVerified=true&returnKeystoreEntries=false"
	certificateUpdateQuery = certificateImportQuery + "&update=true"
)

func certificateResourcePath(hexAlias, query string) string {
	return v2.BuildPath(certificateResourcesEntitySet, v2.KeyPredicate(hexAlias), "") + "/$value?" + query
}

func certificateValuePath(hexAlias string) string {
	return v2.BuildPath(keystoreEntriesEntitySet, v2.KeyPredicate(hexAlias), "") + "/Certificate/$value"
}

// ImportCertificate adds a certificate under a new alias, sending its raw
// PEM bytes exactly as provided (this client neither reformats nor re-wraps
// the content): PUT /api/v1/CertificateResources('{Hexalias}')/$value. SAP's
// own documentation flags the PUT-creates-an-entity quirk: "Although an
// HTTP PUT method is used, a new OData entity is created with the sample
// request." SAP's documented example body is enclosed in square brackets;
// the tenant test imported plain PEM without them, so they are
// documentation formatting.
//
// The request confirms the fingerprint (see certificateImportQuery): a
// certificate in a Terraform configuration is one the configuration's
// author chose to trust. An alias that already exists is rejected by SAP.
func (c *Client) ImportCertificate(ctx context.Context, alias string, pemContent []byte) error {
	hexAlias := v2.EncodeUTF8Hex(alias)
	_, err := c.odata.PutRaw(ctx, certificateResourcePath(hexAlias, certificateImportQuery), certificatePEMContentType, pemContent)
	return err
}

// UpdateCertificate replaces the certificate stored under an existing
// alias; it is ImportCertificate with update=true.
func (c *Client) UpdateCertificate(ctx context.Context, alias string, pemContent []byte) error {
	hexAlias := v2.EncodeUTF8Hex(alias)
	_, err := c.odata.PutRaw(ctx, certificateResourcePath(hexAlias, certificateUpdateQuery), certificatePEMContentType, pemContent)
	return err
}

// GetCertificate exports a keystore entry's certificate in PEM form.
// Confirmed directly from SAP's own "Export Certificate" documentation:
// GET /api/v1/KeystoreEntries('{Hexalias}')/Certificate/$value, response
// content type application/pkix-cert.
func (c *Client) GetCertificate(ctx context.Context, alias string) ([]byte, error) {
	hexAlias := v2.EncodeUTF8Hex(alias)
	return c.odata.Get(ctx, certificateValuePath(hexAlias))
}

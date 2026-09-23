package securitycontent

import (
	"context"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const (
	certificateResourcesEntitySet = "CertificateResources"

	// certificatePEMContentType is sent as the request Content-Type when
	// importing or updating a certificate. SAP's own documented GET example
	// (Export Certificate) states the response content type is
	// "application/pkix-cert", but does not separately confirm the request
	// Content-Type for the PUT side of the same endpoint; this client uses
	// the same value on write, on the basis that a $value raw-media-stream
	// endpoint's request and response representations are ordinarily the
	// same media type, not an independently confirmed fact.
	certificatePEMContentType = "application/pkix-cert"
)

func certificateResourcePath(hexAlias string) string {
	return v2.BuildPath(certificateResourcesEntitySet, v2.KeyPredicate(hexAlias), "") + "/$value"
}

func certificateValuePath(hexAlias string) string {
	return v2.BuildPath(keystoreEntriesEntitySet, v2.KeyPredicate(hexAlias), "") + "/Certificate/$value"
}

// PutCertificate imports a new certificate, or updates an existing one at
// the same alias, sending its raw PEM bytes exactly as provided (this
// client neither reformats nor re-wraps the content). Confirmed directly
// from SAP's own "Import and Update Certificate" documentation: PUT
// /api/v1/CertificateResources('{Hexalias}')/$value — SAP's own
// documentation explicitly flags the PUT-creates-an-entity quirk: "Although
// an HTTP PUT method is used, a new OData entity is created with the
// sample request." SAP's documented example body is enclosed in literal
// square brackets ("[-----BEGIN CERTIFICATE-----\n...\n-----END
// CERTIFICATE-----]"); this project could not confirm whether those
// brackets are part of the literal wire content or SAP's own
// documentation-authoring convention for "insert content here" (no other
// example on any page this project reviewed uses that convention for an
// inline value), and treats them as documentation formatting, not literal
// bytes to send — see docs/guides/security-content.md.
func (c *Client) PutCertificate(ctx context.Context, alias string, pemContent []byte) error {
	hexAlias := v2.EncodeUTF8Hex(alias)
	_, err := c.odata.PutRaw(ctx, certificateResourcePath(hexAlias), certificatePEMContentType, pemContent)
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

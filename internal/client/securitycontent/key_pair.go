package securitycontent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const keyPairGenerationRequestsEntitySet = "KeyPairGenerationRequests"

// KeyPairGenerationRequest is the write model for generating a new,
// SAP-managed key pair. Every field, its mandatory/optional status, type,
// and default value are confirmed verbatim from SAP's own "Generate a Key
// Pair" documentation's Input Properties table and its worked request
// example. The private key material never leaves SAP: this type (and every
// type in this file) has no field for it, and no operation in this client
// ever requests one.
type KeyPairGenerationRequest struct {
	// Alias is the keystore entry alias to create. Mandatory.
	Alias string
	// KeyType is one of "RSA" (default), "DSA", "EC" — confirmed as the
	// complete, fixed enum SAP documents for this field.
	KeyType string
	// SignatureAlgorithm's valid values depend on KeyType, confirmed
	// verbatim: RSA -> SHA-512/RSA (default), SHA-256/RSA, SHA-384/RSA,
	// SHA-224/RSA, SHA-1/RSA; DSA -> SHA-256/DSA, SHA-224/DSA, SHA-1/DSA;
	// EC -> SHA-512/ECDSA, SHA-256/ECDSA, SHA-1/ECDSA. SAP's own worked
	// example inconsistently writes "SHA256/RSA" (no hyphen) where its own
	// field-table enum for the same value is "SHA-256/RSA" (with a
	// hyphen); this client always sends the hyphenated form from the enum
	// table, treating the un-hyphenated example as a documentation
	// artifact, not a second valid wire value — see
	// docs/guides/security-content.md.
	SignatureAlgorithm string
	// KeySize must be specified for KeyType RSA/DSA. For KeyType EC,
	// either KeySize or KeyAlgorithmParameter must be specified, and
	// KeySize (if used) must be 112-571. 0 means "not specified" (omitted
	// from the request).
	KeySize int
	// KeyAlgorithmParameter is only used for KeyType "EC": the standard
	// name of an Elliptic Curve domain parameter (for example "secp256r1",
	// "NIST P-256"), used instead of KeySize. Empty means "not specified".
	KeyAlgorithmParameter string
	// CommonName is mandatory: the certificate subject DN's CN component.
	CommonName string
	// OrganizationUnit, Organization, Locality, State, Email are optional
	// subject DN components.
	OrganizationUnit string
	Organization     string
	Locality         string
	State            string
	Email            string
	// Country is mandatory: SAP documents "two characters required."
	Country string
	// ValidNotBefore/ValidNotAfter are optional; SAP applies its own
	// defaults (current time / current time + 3 years) when left unset.
	// nil means "not specified" (omitted from the request, letting SAP's
	// default apply) rather than the Go zero time.
	ValidNotBefore *time.Time
	ValidNotAfter  *time.Time
}

// keyPairGenerationWireRequest is the exact JSON shape SAP documents for
// POST /KeyPairGenerationRequests, field for field including Hexalias —
// SAP's own documentation states Hexalias is mandatory but its value is
// "Any dummy hex value: System calculates the correct hex value from the
// specified Alias property," so this client always sends a fixed,
// deterministic placeholder rather than exposing this transport detail to
// a caller who would otherwise have to invent one.
type keyPairGenerationWireRequest struct {
	Hexalias              string  `json:"Hexalias"`
	Alias                 string  `json:"Alias"`
	KeyType               string  `json:"KeyType,omitempty"`
	SignatureAlgorithm    string  `json:"SignatureAlgorithm,omitempty"`
	KeySize               int     `json:"KeySize,omitempty"`
	KeyAlgorithmParameter string  `json:"KeyAlgorithmParameter,omitempty"`
	CommonName            string  `json:"CommonName"`
	OrganizationUnit      string  `json:"OrganizationUnit,omitempty"`
	Organization          string  `json:"Organization,omitempty"`
	Locality              string  `json:"Locality,omitempty"`
	State                 string  `json:"State,omitempty"`
	Country               string  `json:"Country"`
	Email                 string  `json:"Email,omitempty"`
	ValidNotBefore        *string `json:"ValidNotBefore,omitempty"`
	ValidNotAfter         *string `json:"ValidNotAfter,omitempty"`
}

// keyPairGenerationDummyHexalias is sent as the required-but-ignored
// Hexalias property on every generation request. SAP's own documentation
// states the system recalculates the correct value from Alias regardless
// of what is sent, so any well-formed hex string is equally valid; this
// one is fixed and deterministic so this client's requests are themselves
// deterministic.
const keyPairGenerationDummyHexalias = "00"

// GenerateKeyPair creates a new SAP-generated key pair as a persistent
// tenant keystore entry. Confirmed directly from SAP's own "Generate a Key
// Pair" documentation: POST /api/v1/KeyPairGenerationRequests. The
// generation request itself is a one-shot command, not a persistent
// object — read the resulting keystore entry back with GetKeystoreEntry
// (by req.Alias), not by trying to re-read this request.
func (c *Client) GenerateKeyPair(ctx context.Context, req KeyPairGenerationRequest) error {
	wire := keyPairGenerationWireRequest{
		Hexalias:              keyPairGenerationDummyHexalias,
		Alias:                 req.Alias,
		KeyType:               req.KeyType,
		SignatureAlgorithm:    req.SignatureAlgorithm,
		KeySize:               req.KeySize,
		KeyAlgorithmParameter: req.KeyAlgorithmParameter,
		CommonName:            req.CommonName,
		OrganizationUnit:      req.OrganizationUnit,
		Organization:          req.Organization,
		Locality:              req.Locality,
		State:                 req.State,
		Country:               req.Country,
		Email:                 req.Email,
	}
	if req.ValidNotBefore != nil {
		lit := v2.FormatDateLiteral(*req.ValidNotBefore)
		wire.ValidNotBefore = &lit
	}
	if req.ValidNotAfter != nil {
		lit := v2.FormatDateLiteral(*req.ValidNotAfter)
		wire.ValidNotAfter = &lit
	}

	payload, err := json.Marshal(wire)
	if err != nil {
		return fmt.Errorf("securitycontent: encoding key pair generation request: %w", err)
	}

	_, err = c.odata.Post(ctx, keyPairGenerationRequestsEntitySet, payload)
	return err
}

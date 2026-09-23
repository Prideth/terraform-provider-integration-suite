package partnerdirectory

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const alternativePartnersEntitySet = "AlternativePartners"

// AlternativePartner is the wire representation of an AlternativePartners
// entity: a mapping from an external identity tuple (Agency, Scheme, Id) to
// an internal Partner ID (Pid). SAP's OData entity key is not the plain
// Agency/Scheme/Id strings themselves but their hex-encoded form
// (Hexagency/Hexscheme/Hexid) — confirmed by a documented example request
// URL, "AlternativePartners(Hexagency='6167656e637931',...)", where
// "6167656e637931" is the lowercase hex encoding of the UTF-8 bytes of
// "agency1". This client hides that transport detail: HexAgency/HexScheme/
// HexId are computed by EncodeAlternativePartnerKey and used only to build
// request paths, never exposed as something a caller sets directly.
type AlternativePartner struct {
	Agency string `json:"Agency"`
	Scheme string `json:"Scheme"`
	Id     string `json:"Id"`
	Pid    string `json:"Pid"`
}

// EncodeAlternativePartnerKey deterministically converts an
// (agency, scheme, id) tuple to the hex-encoded key components
// (Hexagency, Hexscheme, Hexid) SAP's AlternativePartners entity uses as
// its actual OData key. The encoding is the lowercase hex representation
// of each string's UTF-8 bytes — plain encoding/hex.EncodeToString,
// matching SAP's own documented example. This makes the transform total
// (every possible string, including one containing spaces, punctuation, or
// non-ASCII Unicode, has exactly one encoding) and trivially reversible,
// which is what lets AlternativePartnerImportID below build an
// unambiguous Terraform import identifier out of three strings that may
// themselves contain "/" or any other delimiter.
func EncodeAlternativePartnerKey(agency, scheme, id string) (hexAgency, hexScheme, hexID string) {
	return v2.EncodeUTF8Hex(agency), v2.EncodeUTF8Hex(scheme), v2.EncodeUTF8Hex(id)
}

// DecodeAlternativePartnerKeyComponent reverses one component of
// EncodeAlternativePartnerKey. It is used to recover the plain agency,
// scheme, or id string from a Terraform import ID built out of hex
// components (see AlternativePartnerImportID/ParseAlternativePartnerImportID
// in the provider layer), and to validate that a hex string this client
// did not itself produce (for example one a practitioner typed by hand) is
// well-formed before it is used to build a request path.
func DecodeAlternativePartnerKeyComponent(hexValue string) (string, error) {
	return v2.DecodeUTF8Hex(hexValue)
}

func alternativePartnerKey(agency, scheme, id string) (string, error) {
	hexAgency, hexScheme, hexID := EncodeAlternativePartnerKey(agency, scheme, id)
	return v2.CompositeKeyPredicate("Hexagency", hexAgency, "Hexscheme", hexScheme, "Hexid", hexID)
}

// GetAlternativePartner reads a single alternative partner mapping by its
// (agency, scheme, id) tuple, encoding it to the hex key SAP's API expects.
func (c *Client) GetAlternativePartner(ctx context.Context, agency, scheme, id string) (*AlternativePartner, error) {
	key, err := alternativePartnerKey(agency, scheme, id)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Get(ctx, v2.BuildPath(alternativePartnersEntitySet, key, ""))
	if err != nil {
		return nil, err
	}

	var ap AlternativePartner
	if err := v2.DecodeEntity(body, &ap); err != nil {
		return nil, err
	}
	return &ap, nil
}

// CreateAlternativePartner creates a new alternative partner mapping. The
// request body carries the plain Agency/Scheme/Id/Pid fields; SAP computes
// and stores the hex key components itself.
func (c *Client) CreateAlternativePartner(ctx context.Context, ap AlternativePartner) (*AlternativePartner, error) {
	payload, err := json.Marshal(ap)
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: encoding alternative partner: %w", err)
	}

	body, err := c.odata.Post(ctx, alternativePartnersEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created AlternativePartner
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateAlternativePartner repoints an existing (agency, scheme, id) mapping
// at a different Pid via PUT, addressed by the hex key. Agency/Scheme/Id
// are the entity's identity and are never sent in the body.
func (c *Client) UpdateAlternativePartner(ctx context.Context, agency, scheme, id, pid string) error {
	key, err := alternativePartnerKey(agency, scheme, id)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(struct {
		Pid string `json:"Pid"`
	}{Pid: pid})
	if err != nil {
		return fmt.Errorf("partnerdirectory: encoding alternative partner: %w", err)
	}

	_, err = c.odata.Put(ctx, v2.BuildPath(alternativePartnersEntitySet, key, ""), payload)
	return err
}

// DeleteAlternativePartner deletes a single alternative partner mapping by
// its (agency, scheme, id) tuple.
func (c *Client) DeleteAlternativePartner(ctx context.Context, agency, scheme, id string) error {
	key, err := alternativePartnerKey(agency, scheme, id)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, v2.BuildPath(alternativePartnersEntitySet, key, ""))
}

// ListAlternativePartners returns every alternative partner mapping for a
// given internal Pid, following SAP's server-driven paging until
// exhausted.
func (c *Client) ListAlternativePartners(ctx context.Context, pid string) ([]AlternativePartner, error) {
	filter := v2.Query{Filter: v2.FilterEquals("Pid", pid)}
	path := v2.BuildPath(alternativePartnersEntitySet, "", filter.Encode())

	partners, err := v2.GetAllPages[AlternativePartner](ctx, c.odata, path)
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: listing alternative partners: %w", err)
	}
	return partners, nil
}

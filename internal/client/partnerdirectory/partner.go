package partnerdirectory

import (
	"context"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const partnersEntitySet = "Partners"

// Partner is the wire representation of a Partners entity: the read-only
// view of one Partner ID (Pid) known to the Partner Directory.
//
// Partners has no confirmed public create operation: a Pid comes into
// existence implicitly the first time a caller creates a child entity
// (a StringParameter, BinaryParameter, AlternativePartner, AuthorizedUser,
// or UserCredentialParameter) referencing it, and SAP's own documentation
// says PID uniqueness "is ensured by the tenant owner application" — i.e.
// the caller picks the value, there is no server-side allocation to model
// as a Terraform Create. This is why this provider exposes Partners only
// through read-only data sources (data.sapintegrationsuite_partner,
// data.sapintegrationsuite_partners), never as a resource: a Terraform
// resource needs a safe declarative Create/Delete lifecycle, and Partners
// has neither a confirmed Create nor a safe Delete (SAP documents deleting
// a Pid as removing every child entry belonging to it in one call, which
// would make a Partner resource's Destroy capable of erasing content
// owned by entirely different Terraform resources or modules).
type Partner struct {
	Pid string `json:"Pid"`
}

// GetPartner reads a single partner by its Pid. This only confirms the Pid
// is known to the Partner Directory; Partners carries no other documented
// properties.
func (c *Client) GetPartner(ctx context.Context, pid string) (*Partner, error) {
	path := v2.BuildPath(partnersEntitySet, v2.KeyPredicate(pid), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var p Partner
	if err := v2.DecodeEntity(body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListPartners returns every partner ID known to the Partner Directory,
// following SAP's server-driven paging until exhausted. This is the
// primary brownfield discovery mechanism for this provider's Partner
// Directory support: since Partners is not a resource, a practitioner
// adopting existing Partner Directory content needs a way to find out
// which Pid values already exist.
func (c *Client) ListPartners(ctx context.Context) ([]Partner, error) {
	partners, err := v2.GetAllPages[Partner](ctx, c.odata, partnersEntitySet)
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: listing partners: %w", err)
	}
	return partners, nil
}

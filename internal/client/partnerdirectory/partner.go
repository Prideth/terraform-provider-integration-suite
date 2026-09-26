package partnerdirectory

import (
	"context"
	"fmt"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const partnersEntitySet = "Partners"

// Partner is the wire representation of a Partners entity: the read-only
// view of one Partner ID (Pid) known to the Partner Directory.
//
// Partners has no public create operation: a Pid comes into
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

// GetPartner looks up a single partner by its Pid. This only confirms the
// Pid is known to the Partner Directory; Partners carries no other
// properties. SAP refuses a read by key ("Reading of single partner
// entities is not supported", 400, tenant test September 2026), so the
// lookup filters the collection by Pid, which the same test answered with
// 200. A Pid with no entries is reported as a 404 *apierror.Error.
func (c *Client) GetPartner(ctx context.Context, pid string) (*Partner, error) {
	query := v2.Query{Filter: v2.FilterEquals("Pid", pid)}.Encode()

	body, err := c.odata.Get(ctx, v2.BuildPath(partnersEntitySet, "", query))
	if err != nil {
		return nil, err
	}

	var partners []Partner
	if err := v2.DecodeCollection(body, &partners); err != nil {
		return nil, err
	}
	for i := range partners {
		if partners[i].Pid == pid {
			return &partners[i], nil
		}
	}
	return nil, &apierror.Error{StatusCode: 404, Message: "partner " + pid + " is not in the Partner Directory"}
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

package securitycontent

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const secureParametersEntitySet = "SecureParameters"

// SecureParameter is the read view of a SecureParameters entity (Security
// Content API): a confidential value deployed under an alias, for example
// for custom adapters. It deliberately has no field for the secret. A tenant
// returned SecureParam as null on every read (September 2026), and this
// client never decodes it, so the value cannot reach state or a log line.
//
// SAP Help documents the artifact only in the Monitor UI; the entity set and
// its properties come from the tenant $metadata, and create (POST), read
// (GET by Name), update (PUT) and delete were verified on a tenant.
type SecureParameter struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
	DeployedBy  string `json:"DeployedBy"`
	DeployedOn  string `json:"DeployedOn"`
	Status      string `json:"Status"`
}

// secureParameterWriteRequest is the body for create and update. It is only
// ever marshaled, never a target of json.Unmarshal.
type secureParameterWriteRequest struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
	SecureParam string `json:"SecureParam"`
}

func secureParameterPath(name string) string {
	return v2.BuildPath(secureParametersEntitySet, v2.KeyPredicate(name), "")
}

// GetSecureParameter reads a secure parameter's metadata by name. Like the
// other Security Content entity sets it is requested without query options,
// which the service rejects.
func (c *Client) GetSecureParameter(ctx context.Context, name string) (*SecureParameter, error) {
	body, err := c.odata.Get(ctx, secureParameterPath(name))
	if err != nil {
		return nil, err
	}
	var sp SecureParameter
	if err := v2.DecodeEntity(body, &sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// CreateSecureParameter deploys a new secure parameter. A tenant answered the
// POST with 202 and no body, so the entry is read back when the response is
// empty.
func (c *Client) CreateSecureParameter(ctx context.Context, name, description, value string) (*SecureParameter, error) {
	payload, err := json.Marshal(secureParameterWriteRequest{Name: name, Description: description, SecureParam: value}) //nolint:gosec // G117: the secret is sent to SAP's create API on purpose; it never reaches state or a log line
	if err != nil {
		return nil, fmt.Errorf("securitycontent: encoding secure parameter: %w", err)
	}

	body, err := c.odata.Post(ctx, secureParametersEntitySet, payload)
	if err != nil {
		return nil, err
	}
	if v2.EmptyBody(body) {
		return c.GetSecureParameter(ctx, name)
	}
	var created SecureParameter
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateSecureParameter redeploys a secure parameter with PUT. The value is
// sent on every update, as the Monitor UI requires re-entering it; a tenant
// accepted such a PUT with 202.
func (c *Client) UpdateSecureParameter(ctx context.Context, name, description, value string) error {
	payload, err := json.Marshal(secureParameterWriteRequest{Name: name, Description: description, SecureParam: value}) //nolint:gosec // G117: see CreateSecureParameter
	if err != nil {
		return fmt.Errorf("securitycontent: encoding secure parameter: %w", err)
	}
	_, err = c.odata.Put(ctx, secureParameterPath(name), payload)
	return err
}

// DeleteSecureParameter deletes a secure parameter by name. A tenant
// answered 202 and a read afterwards returned 404.
func (c *Client) DeleteSecureParameter(ctx context.Context, name string) error {
	return c.odata.Delete(ctx, secureParameterPath(name))
}

// Length limits from the tenant $metadata (MaxLength of Name and
// SecureParam), which match the 4096 characters SAP's UI page gives for
// Cloud Foundry.
const (
	// MaxSecureParameterNameLength is the longest name SAP accepts.
	MaxSecureParameterNameLength = 150

	// MaxSecureParameterValueLength is the longest secret SAP stores.
	MaxSecureParameterValueLength = 4096
)

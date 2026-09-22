package partnerdirectory

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const userCredentialParametersEntitySet = "UserCredentialParameters"

// UserCredentialParameter is the READ/IDENTITY-ONLY wire representation of
// a UserCredentialParameters entity: deliberately, this struct has no
// Password field. Whatever SAP's GET (or a create/update response) does or
// does not return for that property, this client never decodes it into a
// Go value, so a password can never end up copied into Terraform state,
// a diagnostic, or a log line by way of this struct — the safety property
// holds regardless of SAP's actual response shape, which this project could
// not fully confirm. See createUserCredentialParameterRequest below for the
// separate, write-only request shape used to submit a password.
type UserCredentialParameter struct {
	Pid  string `json:"Pid"`
	Id   string `json:"Id"`
	User string `json:"User"`
}

// createUserCredentialParameterRequest is the request body shape for
// creating a user credential parameter: Pid, Id, User, and Password, per
// SAP's documented example. This type exists only to be marshaled — it is
// never a target of json.Unmarshal, so there is no code path that could
// accidentally decode a password out of an API response into it.
type createUserCredentialParameterRequest struct {
	Pid      string `json:"Pid"`
	Id       string `json:"Id"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

// GetUserCredentialParameter reads a single user credential parameter's
// identity (Pid, Id, User) by its (Pid, Id) key. It never requests or
// decodes a password.
func (c *Client) GetUserCredentialParameter(ctx context.Context, pid, id string) (*UserCredentialParameter, error) {
	key, err := v2.CompositeKeyPredicate("Pid", pid, "Id", id)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Get(ctx, v2.BuildPath(userCredentialParametersEntitySet, key, ""))
	if err != nil {
		return nil, err
	}

	var ucp UserCredentialParameter
	if err := v2.DecodeEntity(body, &ucp); err != nil {
		return nil, err
	}
	return &ucp, nil
}

// CreateUserCredentialParameter creates a new user credential parameter.
// password is sent once, in this single request, and is never returned:
// the response is decoded into UserCredentialParameter, whose type has no
// Password field to receive it even if SAP's response body happened to
// include one.
//
// SAP documents UserCredentialParameter and CertificateUserMapping as
// unable to be combined with other entity types in an OData ChangeSet
// (batch) request; this client only ever issues it as a standalone
// request, so that constraint does not need to be enforced here, but it is
// the reason this entity is never grouped with any other Partner Directory
// write in a single call.
func (c *Client) CreateUserCredentialParameter(ctx context.Context, pid, id, user, password string) (*UserCredentialParameter, error) {
	payload, err := json.Marshal(createUserCredentialParameterRequest{Pid: pid, Id: id, User: user, Password: password})
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: encoding user credential parameter: %w", err)
	}

	body, err := c.odata.Post(ctx, userCredentialParametersEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created UserCredentialParameter
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteUserCredentialParameter deletes a single user credential parameter
// by its (Pid, Id) key.
//
// There is no UpdateUserCredentialParameter: no public documentation of an
// in-place PUT/PATCH for this entity's password was found during this
// feature's research pass, and guessing at one for a security-sensitive
// credential is not an acceptable risk. Rotating a credential's password
// is modeled as replacing the resource (delete, then create) — see
// resource_partner_user_credential_parameter.go — using only the two
// operations this client can confirm.
func (c *Client) DeleteUserCredentialParameter(ctx context.Context, pid, id string) error {
	key, err := v2.CompositeKeyPredicate("Pid", pid, "Id", id)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, v2.BuildPath(userCredentialParametersEntitySet, key, ""))
}

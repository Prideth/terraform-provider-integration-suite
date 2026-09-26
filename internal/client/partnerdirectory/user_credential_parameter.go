package partnerdirectory

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const userCredentialParametersEntitySet = "UserCredentialParameters" // #nosec G101 -- an OData entity set name, not a credential value

// UserCredentialParameter is the READ/IDENTITY-ONLY wire representation of
// a UserCredentialParameters entity: deliberately, this struct has no
// Password field. SAP returns Password as null (or, with the query option
// returnHashedPassword=SHA256, as a hash this client never requests), and
// this client never decodes it into a Go value, so a password can never
// end up copied into Terraform state, a diagnostic, or a log line by way of
// this struct. See createUserCredentialParameterRequest below for the
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

// CreateUserCredentialParameter creates a user credential parameter with
// POST. SAP's POST is an upsert: "You can also use the POST request to
// update a User Credentials parameter with the same values for PID and Id",
// so a POST for an existing (Pid, Id) overwrites it. Callers that must not
// overwrite check for an existing entry first.
//
// password is sent once, in this single request, and is never returned:
// the response is decoded into UserCredentialParameter, whose type has no
// Password field to receive it even if SAP's response body happened to
// include one.
//
// SAP allows a UserCredentialParameter request only alone in an OData
// change set; this client never batches it.
func (c *Client) CreateUserCredentialParameter(ctx context.Context, pid, id, user, password string) (*UserCredentialParameter, error) {
	payload, err := json.Marshal(createUserCredentialParameterRequest{Pid: pid, Id: id, User: user, Password: password}) //nolint:gosec // G117: deliberately marshals the password into the request body sent to SAP's Create API -- that is the whole purpose of this call, not a leak; see resource_partner_user_credential_parameter.go for why it never reaches Terraform state or a log line
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

// UpdateUserCredentialParameter changes the user and password of an
// existing user credential parameter. SAP documents that PUT is not
// supported for this entity and that POST with the same Pid and Id updates
// it, so this sends the same POST as CreateUserCredentialParameter.
func (c *Client) UpdateUserCredentialParameter(ctx context.Context, pid, id, user, password string) (*UserCredentialParameter, error) {
	return c.CreateUserCredentialParameter(ctx, pid, id, user, password)
}

// DeleteUserCredentialParameter deletes a single user credential parameter
// by its (Pid, Id) key.
func (c *Client) DeleteUserCredentialParameter(ctx context.Context, pid, id string) error {
	key, err := v2.CompositeKeyPredicate("Pid", pid, "Id", id)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, v2.BuildPath(userCredentialParametersEntitySet, key, ""))
}

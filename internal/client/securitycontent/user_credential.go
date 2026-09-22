package securitycontent

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const userCredentialsEntitySet = "UserCredentials"

// UserCredential is the READ/IDENTITY-ONLY wire representation of a
// UserCredentials entity (Security Content API): deliberately, this struct
// has no Password field. Whatever SAP's GET (or a create/update response)
// does or does not return for that property, this client never decodes it
// into a Go value, so a password can never end up copied into Terraform
// state, a diagnostic, or a log line by way of this struct.
//
// Name is both the artifact's display name and its OData key: SAP's own UI
// documentation states "The artifact name is used as an alias for the
// confidential data assigned by this parameter", so this client treats it
// as the sole identifier, the same as the alias an integration flow adapter
// references.
//
// Kind and CompanyId model the credential "Type" SAP's UI exposes
// (generic Basic, SuccessFactors, OpenConnectors): CompanyId is only
// meaningful when Kind is "SuccessFactors". Their exact OData property
// casing is corroborated by a documented example payload but not by this
// project's own inspection of a live tenant's $metadata; see
// docs/guides/security-content.md for the verification status of every
// field on this type.
type UserCredential struct {
	Name        string `json:"Name"`
	Kind        string `json:"Kind,omitempty"`
	Description string `json:"Description,omitempty"`
	User        string `json:"User"`
	CompanyID   string `json:"CompanyId,omitempty"`
}

// userCredentialWriteRequest is the request body shape for creating or
// redeploying (editing) a user credential artifact: Name, Kind,
// Description, User, Password, and CompanyId, per SAP's documented example
// payload. This type exists only to be marshaled — it is never a target of
// json.Unmarshal, so there is no code path that could accidentally decode a
// password out of an API response into it.
type userCredentialWriteRequest struct {
	Name        string `json:"Name"`
	Kind        string `json:"Kind,omitempty"`
	Description string `json:"Description,omitempty"`
	User        string `json:"User"`
	Password    string `json:"Password"`
	CompanyID   string `json:"CompanyId,omitempty"`
}

func userCredentialPath(name string) string {
	return v2.BuildPath(userCredentialsEntitySet, v2.KeyPredicate(name), "")
}

// GetUserCredential reads a single user credential artifact's identity and
// metadata by its Name (alias). It never requests or decodes a password.
func (c *Client) GetUserCredential(ctx context.Context, name string) (*UserCredential, error) {
	body, err := c.odata.Get(ctx, userCredentialPath(name))
	if err != nil {
		return nil, err
	}

	var cred UserCredential
	if err := v2.DecodeEntity(body, &cred); err != nil {
		return nil, err
	}
	return &cred, nil
}

// CreateUserCredential creates (deploys) a new user credential artifact.
// password is sent once, in this single request, and is never returned:
// the response is decoded into UserCredential, whose type has no Password
// field to receive it even if SAP's response body happened to include one.
func (c *Client) CreateUserCredential(ctx context.Context, cred UserCredential, password string) (*UserCredential, error) {
	payload, err := json.Marshal(userCredentialWriteRequest{
		Name:        cred.Name,
		Kind:        cred.Kind,
		Description: cred.Description,
		User:        cred.User,
		Password:    password,
		CompanyID:   cred.CompanyID,
	})
	if err != nil {
		return nil, fmt.Errorf("securitycontent: encoding user credential: %w", err)
	}

	body, err := c.odata.Post(ctx, userCredentialsEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created UserCredential
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateUserCredential redeploys (edits) an existing user credential
// artifact in place via PUT, replacing every mutable field at once. SAP's
// Manage Security Material UI documents "Edit" as a supported action for
// Credentials artifacts ("You can also edit and redeploy an existing
// artifact"), so this provider models a password_wo_version change (or any
// other mutable-field change) as an in-place update rather than a
// replacement. password must be resupplied on every edit: SAP documents
// this explicitly for the sibling OAuth2 Client Credentials artifact
// ("Every time you edit ... you must re-enter the Client Secret"), and this
// client assumes the same holds for User Credentials since neither
// artifact type documents returning a stored secret for a partial edit.
//
// The response body is deliberately not decoded here: like
// UpdateAccessPolicy, this project has not confirmed whether a successful
// PUT returns the updated entity or 204 No Content, so the caller re-reads
// the entity with GetUserCredential instead of trusting this call's
// response shape.
func (c *Client) UpdateUserCredential(ctx context.Context, cred UserCredential, password string) error {
	payload, err := json.Marshal(userCredentialWriteRequest{
		Name:        cred.Name,
		Kind:        cred.Kind,
		Description: cred.Description,
		User:        cred.User,
		Password:    password,
		CompanyID:   cred.CompanyID,
	})
	if err != nil {
		return fmt.Errorf("securitycontent: encoding user credential: %w", err)
	}

	_, err = c.odata.Put(ctx, userCredentialPath(cred.Name), payload)
	return err
}

// DeleteUserCredential deletes a user credential artifact by its Name.
func (c *Client) DeleteUserCredential(ctx context.Context, name string) error {
	return c.odata.Delete(ctx, userCredentialPath(name))
}

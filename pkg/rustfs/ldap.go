package rustfs

import (
	"context"
	"encoding/json"
)

// CreateLDAPServiceAccount creates a service account scoped to an LDAP user.
// The request body mirrors the standard ServiceAccount struct with TargetUser
// carrying the LDAP user (distinguished name) the account is scoped to.
//
// RustFS has no LDAP directory backend, so the target user must already exist
// in the identity store (e.g. provisioned by a configured LDAP IdP), otherwise
// the server rejects the request with "target user not exist".
func (c *RustfsAdmin) CreateLDAPServiceAccount(account ServiceAccount) error {
	normalizeServiceAccount(&account)
	//#nosec G117 — AccessKey is a public identifier, not a secret
	bytes, err := json.Marshal(account)
	if err != nil {
		return err
	}
	req_data := RequestData{
		Method:  "PUT",
		RelPath: "idp/ldap/add-service-account",
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = c.doRequest(ctx, req_data)
	return err
}

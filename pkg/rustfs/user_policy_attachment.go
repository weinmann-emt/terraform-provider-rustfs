package rustfs

import (
	"context"
	"encoding/json"
)

// userPolicyAssociationRequest is the JSON body accepted by the RustFS
// builtin-policy attach/detach endpoints. Exactly one of User or Group must
// be set, together with at least one policy name.
type userPolicyAssociationRequest struct {
	User     string   `json:"user,omitempty"`
	Group    string   `json:"group,omitempty"`
	Policies []string `json:"policies"`
}

// AttachUserPolicy attaches a canned policy to a user, leaving any policies
// the user already holds untouched. RustFS stores a user's policies as a
// comma-separated list, so this appends to the existing list instead of
// replacing it (unlike PUT set-user-or-group-policy, which replaces the
// whole mapping and is used for the single primary policy on rustfs_user).
func (c *RustfsAdmin) AttachUserPolicy(user, policy string) error {
	return c.userPolicyAssociation(user, policy, "idp/builtin/policy/attach")
}

// DetachUserPolicy detaches a single canned policy from a user, leaving any
// other attached policies untouched. The RustFS admin API exposes no DELETE
// variant of set-user-or-group-policy, so removal goes through the dedicated
// POST idp/builtin/policy/detach endpoint.
func (c *RustfsAdmin) DetachUserPolicy(user, policy string) error {
	return c.userPolicyAssociation(user, policy, "idp/builtin/policy/detach")
}

func (c *RustfsAdmin) userPolicyAssociation(user, policy, relPath string) error {
	body, err := json.Marshal(userPolicyAssociationRequest{
		User:     user,
		Policies: []string{policy},
	})
	if err != nil {
		return err
	}

	reqData := RequestData{
		Method:  "POST",
		RelPath: relPath,
		Content: body,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = c.doRequest(ctx, reqData)
	return err
}

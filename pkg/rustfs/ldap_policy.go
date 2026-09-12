package rustfs

import (
	"context"
	"encoding/json"
)

// LDAPPolicyAttachment identifies a canned policy attached to an LDAP user or
// group. The RustFS admin API expresses the target via a dedicated field, so
// the JSON body carries the policy name under "policies" and the distinguished
// name under either "user" or "group".
type LDAPPolicyAttachment struct {
	UserOrGroup string
	PolicyName  string
	IsGroup     bool
}

func (a LDAPPolicyAttachment) MarshalJSON() ([]byte, error) {
	target := "user"
	if a.IsGroup {
		target = "group"
	}
	payload := map[string]any{
		"policies": []string{a.PolicyName},
		target:     a.UserOrGroup,
	}
	return json.Marshal(payload)
}

// AttachLDAPPolicy attaches a canned policy to an LDAP user or group.
func (c *RustfsAdmin) AttachLDAPPolicy(req LDAPPolicyAttachment) error {
	return c.ldapPolicyOperation("attach", req)
}

// DetachLDAPPolicy detaches a canned policy from an LDAP user or group.
func (c *RustfsAdmin) DetachLDAPPolicy(req LDAPPolicyAttachment) error {
	return c.ldapPolicyOperation("detach", req)
}

func (c *RustfsAdmin) ldapPolicyOperation(operation string, req LDAPPolicyAttachment) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqData := RequestData{
		Method:  "POST",
		RelPath: "idp/ldap/policy/" + operation,
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

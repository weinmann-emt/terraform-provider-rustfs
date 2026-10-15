package rustfs

import (
	"context"
	"net/url"
	"slices"
	"strings"
	"sync"
)

// groupPolicyLocks serialises read-modify-write cycles on a group's policy
// mapping. set-user-or-group-policy replaces the whole mapping, so concurrent
// attachments to the same group (Terraform applies resources in parallel)
// would otherwise overwrite each other.
var (
	groupPolicyLocksMu sync.Mutex
	groupPolicyLocks   = map[string]*sync.Mutex{}
)

func (c *RustfsAdmin) lockGroupPolicy(group string) func() {
	key := c.endpointURL + "\x00" + group
	groupPolicyLocksMu.Lock()
	mu, ok := groupPolicyLocks[key]
	if !ok {
		mu = &sync.Mutex{}
		groupPolicyLocks[key] = mu
	}
	groupPolicyLocksMu.Unlock()
	mu.Lock()
	return mu.Unlock
}

// GetGroupPolicies returns the canned policies currently mapped to a group.
// The admin API reports them as a single comma-separated string.
func (c *RustfsAdmin) GetGroupPolicies(group string) ([]string, error) {
	info, err := c.GetGroup(group)
	if err != nil {
		return nil, err
	}
	return splitPolicies(info.Policy), nil
}

// AttachGroupPolicy adds a canned policy to an IAM group, keeping any
// policies that are already attached.
func (c *RustfsAdmin) AttachGroupPolicy(group, policy string) error {
	unlock := c.lockGroupPolicy(group)
	defer unlock()

	policies, err := c.GetGroupPolicies(group)
	if err != nil {
		return err
	}
	if slices.Contains(policies, policy) {
		return nil
	}
	return c.setGroupPolicies(group, append(policies, policy))
}

// DetachGroupPolicy removes a canned policy from an IAM group, keeping the
// other attached policies. An empty policyName clears the mapping entirely
// (same semantics as MinIO set-policy with an empty name).
func (c *RustfsAdmin) DetachGroupPolicy(group, policy string) error {
	unlock := c.lockGroupPolicy(group)
	defer unlock()

	policies, err := c.GetGroupPolicies(group)
	if err != nil {
		return err
	}
	if !slices.Contains(policies, policy) {
		return nil
	}
	remaining := slices.DeleteFunc(policies, func(p string) bool { return p == policy })
	return c.setGroupPolicies(group, remaining)
}

func (c *RustfsAdmin) setGroupPolicies(group string, policies []string) error {
	query := url.Values{}
	query.Set("userOrGroup", group)
	query.Set("policyName", strings.Join(policies, ","))
	query.Set("isGroup", "true")
	reqData := RequestData{
		Method:      "PUT",
		RelPath:     "set-user-or-group-policy",
		QueryValues: query,
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

func splitPolicies(s string) []string {
	var policies []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			policies = append(policies, p)
		}
	}
	return policies
}

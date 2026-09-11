package rustfs

import (
	"context"
	"encoding/json"
	"fmt"
)

// AuditTargetKeyValue is a single key/value pair in an audit target config.
type AuditTargetKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AuditTarget describes an audit-log target as returned by the admin list
// endpoint. The admin list response does not echo the target configuration
// (endpoint, auth token, ...) — only identity and runtime health metadata, so
// those config fields must be preserved by the caller rather than read back.
type AuditTarget struct {
	AccountID    string `json:"account_id"`
	Service      string `json:"service"`
	Status       string `json:"status"`
	HealthState  string `json:"health_state"`
	HealthReason string `json:"health_reason"`
	Source       string `json:"source"`
}

// auditTargetListResponse is the envelope returned by GET audit/target/list.
type auditTargetListResponse struct {
	AuditEndpoints []AuditTarget `json:"audit_endpoints"`
}

// auditTargetSetBody is the request body for PUT audit/target/{type}/{name}.
type auditTargetSetBody struct {
	KeyValues []AuditTargetKeyValue `json:"key_values"`
}

// ListAuditTargets lists all configured audit-log targets.
func (c *RustfsAdmin) ListAuditTargets() ([]AuditTarget, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "audit/target/list",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out auditTargetListResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.AuditEndpoints, nil
}

// SetAuditTarget adds or edits an audit target, sending keyValues as the
// target configuration. The server validates the keys against the target type
// and automatically enables the target.
func (c *RustfsAdmin) SetAuditTarget(targetType, targetName string, keyValues []AuditTargetKeyValue) error {
	body := auditTargetSetBody{KeyValues: keyValues}
	//#nosec G117 — the marshaled body is the target's own configuration and
	// may contain an auth token, which is expected to be transmitted to the
	// admin API; it is never logged by this client.
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	reqData := RequestData{
		Method:  "PUT",
		RelPath: fmt.Sprintf("audit/target/%s/%s", targetType, targetName),
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

// ResetAuditTarget removes an audit target, reverting it to the server default
// (the admin API exposes no plain delete for targets).
func (c *RustfsAdmin) ResetAuditTarget(targetType, targetName string) error {
	reqData := RequestData{
		Method:  "DELETE",
		RelPath: fmt.Sprintf("audit/target/%s/%s/reset", targetType, targetName),
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

package rustfs

import (
	"context"
	"encoding/json"
)

// SiteReplicationSite models a peer site as accepted by the
// site-replication/add and site-replication/edit admin endpoints. Wire field
// names follow the RustFS (MinIO-compatible) site-replication protocol.
type SiteReplicationSite struct {
	Name          string `json:"name"`
	Endpoint      string `json:"endpoints"`
	AccessKey     string `json:"accessKey,omitempty"`
	SecretKey     string `json:"secretKey,omitempty"`
	SkipTLSVerify bool   `json:"skipTlsVerify,omitempty"`
	CACertPEM     string `json:"caCertPem,omitempty"`
}

// SiteReplicationPeer is a peer site as reported by the site-replication/info
// endpoint.
type SiteReplicationPeer struct {
	Name          string `json:"name"`
	Endpoint      string `json:"endpoint"`
	DeploymentID  string `json:"deploymentID"`
	SkipTLSVerify bool   `json:"skipTlsVerify"`
}

// SiteReplicationInfo is the response of the site-replication/info endpoint.
type SiteReplicationInfo struct {
	Enabled bool                  `json:"enabled"`
	Name    string                `json:"name"`
	Sites   []SiteReplicationPeer `json:"sites"`
}

// SiteReplicationStatus is the response of the add/remove/edit endpoints.
type SiteReplicationStatus struct {
	Status      string `json:"status"`
	ErrorDetail string `json:"errorDetail"`
}

// SiteReplicationResync is the response of the site-replication/resync/op
// endpoint.
type SiteReplicationResync struct {
	OpType string `json:"op"`
	ID     string `json:"id"`
	Status string `json:"status"`
	State  string `json:"state"`
}

func (c *RustfsAdmin) SiteReplicationAdd(site SiteReplicationSite) error {
	//#nosec G117 — AccessKey is a public identifier, not a secret
	body, err := json.Marshal([]SiteReplicationSite{site})
	if err != nil {
		return err
	}
	return c.siteReplicationWrite("site-replication/add", body)
}

func (c *RustfsAdmin) SiteReplicationEdit(site SiteReplicationSite) error {
	//#nosec G117 — AccessKey is a public identifier, not a secret
	body, err := json.Marshal([]SiteReplicationSite{site})
	if err != nil {
		return err
	}
	return c.siteReplicationWrite("site-replication/edit", body)
}

func (c *RustfsAdmin) SiteReplicationRemove(names []string) error {
	body, err := json.Marshal(struct {
		Sites []string `json:"sites"`
	}{Sites: names})
	if err != nil {
		return err
	}
	return c.siteReplicationWrite("site-replication/remove", body)
}

func (c *RustfsAdmin) SiteReplicationInfo() (SiteReplicationInfo, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "site-replication/info",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return SiteReplicationInfo{}, err
	}
	defer resp.Body.Close()
	var info SiteReplicationInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}

// SiteReplicationResyncOp triggers, cancels, or inspects a resync against a
// peer site. operation is one of "start", "cancel", or "status". The server
// identifies the target peer by deploymentID.
func (c *RustfsAdmin) SiteReplicationResyncOp(operation string, peer SiteReplicationPeer) (SiteReplicationResync, error) {
	body, err := json.Marshal(peer)
	if err != nil {
		return SiteReplicationResync{}, err
	}
	reqData := RequestData{
		Method:  "PUT",
		RelPath: "site-replication/resync/op",
		QueryValues: map[string][]string{
			"operation": {operation},
		},
		Content: body,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return SiteReplicationResync{}, err
	}
	defer resp.Body.Close()
	var resync SiteReplicationResync
	err = json.NewDecoder(resp.Body).Decode(&resync)
	return resync, err
}

func (c *RustfsAdmin) siteReplicationWrite(relPath string, body []byte) error {
	reqData := RequestData{
		Method:  "PUT",
		RelPath: relPath,
		Content: body,
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

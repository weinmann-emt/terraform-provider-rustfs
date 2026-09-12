package rustfs

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strings"
)

// RemoteTargetCredentials holds the access credentials for a remote target.
type RemoteTargetCredentials struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// RemoteTarget describes a remote replication target registered on a source
// bucket. The ARN is assigned by the server and never read back from list.
type RemoteTarget struct {
	Arn          string                   `json:"arn,omitempty"`
	Type         string                   `json:"type"`
	Endpoint     string                   `json:"endpoint"`
	Secure       bool                     `json:"secure"`
	Region       string                   `json:"region,omitempty"`
	Path         string                   `json:"path,omitempty"`
	SourceBucket string                   `json:"sourcebucket,omitempty"`
	TargetBucket string                   `json:"targetbucket"`
	Credentials  *RemoteTargetCredentials `json:"credentials,omitempty"`
}

// AddRemoteTarget registers (or, when the ARN is set, updates) a remote target
// on the given source bucket and returns the server-assigned ARN.
func (c *RustfsAdmin) AddRemoteTarget(bucket string, target RemoteTarget) (string, error) {
	body, err := json.Marshal(target)
	if err != nil {
		return "", err
	}
	query := url.Values{"bucket": []string{bucket}}
	// When the caller knows the ARN it is an update: the server only applies a
	// create-path PUT when the target does not already exist (it returns the
	// existing ARN otherwise). To actually persist the new field values the
	// request must be an explicit `update=true` with the `creds` operation group.
	if target.Arn != "" {
		query.Set("update", "true")
		query.Set("creds", "true")
	}
	reqData := RequestData{
		Method:      "PUT",
		RelPath:     "set-remote-target",
		QueryValues: query,
		Content:     body,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.Trim(string(bytes), "\"\n"), nil
}

// ListRemoteTargets returns the remote targets registered on the given source
// bucket.
func (c *RustfsAdmin) ListRemoteTargets(bucket string) ([]RemoteTarget, error) {
	reqData := RequestData{
		Method:      "GET",
		RelPath:     "list-remote-targets",
		QueryValues: url.Values{"bucket": []string{bucket}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var targets []RemoteTarget
	err = json.NewDecoder(resp.Body).Decode(&targets)
	return targets, err
}

// DeleteRemoteTarget removes the remote target identified by arn from the given
// source bucket.
func (c *RustfsAdmin) DeleteRemoteTarget(bucket, arn string) error {
	reqData := RequestData{
		Method:      "DELETE",
		RelPath:     "remove-remote-target",
		QueryValues: url.Values{"bucket": []string{bucket}, "arn": []string{arn}},
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

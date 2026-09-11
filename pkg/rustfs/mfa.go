package rustfs

import (
	"context"
	"encoding/json"
	"net/url"
)

type UserMFAStatus struct {
	AccessKey              string `json:"access_key"`
	Enabled                bool   `json:"enabled"`
	ActivatedAt            string `json:"activated_at,omitempty"`
	RecoveryCodesRemaining int    `json:"recovery_codes_remaining"`
}

// ReadUserMFA returns the MFA status of a user.
func (c *RustfsAdmin) ReadUserMFA(accessKey string) (UserMFAStatus, error) {
	var status UserMFAStatus
	urlValues := make(url.Values)
	urlValues.Set("accessKey", accessKey)
	req_data := RequestData{
		Method:      "GET",
		RelPath:     "user/mfa",
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return status, err
	}
	err = json.NewDecoder(resp.Body).Decode(&status)
	return status, err
}

// ClearUserMFA clears the second factor of a user.
func (c *RustfsAdmin) ClearUserMFA(accessKey string) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", accessKey)
	req_data := RequestData{
		Method:      "DELETE",
		RelPath:     "user/mfa",
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := c.doRequest(ctx, req_data)
	return err
}

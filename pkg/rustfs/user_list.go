package rustfs

import (
	"context"
	"encoding/json"
	"net/url"
)

type UserInfo struct {
	AccessKey string `json:"accessKey"`
	Status    string `json:"status"`
	Policy    string `json:"policyName"`
}

func (c *RustfsAdmin) ListUsers(bucket string) ([]UserInfo, error) {
	query := url.Values{}
	if bucket != "" {
		query.Set("bucket", bucket)
	}
	reqData := RequestData{
		Method:      "GET",
		RelPath:     "list-users",
		QueryValues: query,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var users []UserInfo
	err = json.NewDecoder(resp.Body).Decode(&users)
	return users, err
}

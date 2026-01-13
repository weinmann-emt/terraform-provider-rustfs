package rustfs

import (
	"context"
	"encoding/json"
	"net/url"
)

type UserAccount struct {
	SecretKey string `json:"secretKey"`
	Status    string `json:"status"`
	AccessKey string
	policy    string
	Group     string `json:"-"`
}

type UserInfo struct {
	Status string   `json:"status"`
	Groups []string `json:"memberOf"`
}

func (c RustfsAdmin) CreateUserAccount(user UserAccount) error {

	user.Status = "enabled"
	urlValues := make(url.Values)
	urlValues.Set("accessKey", user.AccessKey)

	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "add-user",
		Content:     bytes,
		QueryValues: urlValues,
	}

	ctx, _ := context.WithCancel(context.Background())
	_, err = c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}

	err = c.addUserToGroup(user.AccessKey, user.Group)
	return err
}

func (c RustfsAdmin) ReadUserAccount(name string) (UserInfo, error) {
	var instance UserInfo
	urlValues := make(url.Values)
	urlValues.Set("accessKey", name)
	req_data := RequestData{
		Method:      "GET",
		RelPath:     "user-info",
		QueryValues: urlValues,
	}
	ctx, _ := context.WithCancel(context.Background())
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return instance, err
	}
	err = json.NewDecoder(resp.Body).Decode(&instance)
	return instance, nil

}

func (c RustfsAdmin) UpdateUserAccount(account UserAccount) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", account.AccessKey)
	urlValues.Set("status", account.Status)
	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "user-info",
		QueryValues: urlValues,
	}
	ctx, _ := context.WithCancel(context.Background())
	_, err := c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return nil
}

func (c RustfsAdmin) DeleteUserAccount(account UserAccount) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", account.AccessKey)
	req_data := RequestData{
		Method:      "DELETE",
		RelPath:     "remove-user",
		QueryValues: urlValues,
	}
	ctx, _ := context.WithCancel(context.Background())
	_, err := c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return err
}

func (c RustfsAdmin) addUserToGroup(user string, group string) error {
	urlValues := make(url.Values)
	urlValues.Set("userOrGroup", user)
	urlValues.Set("policyName", group)
	urlValues.Set("isGroup", "false")
	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "set-user-or-group-policy",
		QueryValues: urlValues,
	}
	ctx, _ := context.WithCancel(context.Background())
	_, err := c.doRequest(ctx, req_data)
	return err
}

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
	Policy    string   `json:"policyName"`
	Groups    []string `json:"memberOf"`
}

func (c *RustfsAdmin) CreateUserAccount(user UserAccount) error {

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}

	if user.Policy != "" {
		return c.addUserToGroup(user.AccessKey, user.Policy)
	}
	return err
}

func (c *RustfsAdmin) ReadUserAccount(name string) (UserAccount, error) {
	var instance UserAccount
	urlValues := make(url.Values)
	urlValues.Set("accessKey", name)
	req_data := RequestData{
		Method:      "GET",
		RelPath:     "user-info",
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return instance, err
	}
	err = json.NewDecoder(resp.Body).Decode(&instance)
	instance.AccessKey = name
	return instance, nil

}

func (c *RustfsAdmin) UpdateUserAccount(account UserAccount) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", account.AccessKey)
	urlValues.Set("status", account.Status)
	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "user-info",
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	if account.Policy != "" {
		return c.addUserToGroup(account.AccessKey, account.Policy)
	}
	return nil
}

func (c *RustfsAdmin) DeleteUserAccount(account UserAccount) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", account.AccessKey)
	req_data := RequestData{
		Method:      "DELETE",
		RelPath:     "remove-user",
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return err
}

func (c *RustfsAdmin) addUserToGroup(user string, group string) error {
	urlValues := make(url.Values)
	urlValues.Set("userOrGroup", user)
	urlValues.Set("policyName", group)
	urlValues.Set("isGroup", "false")
	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "set-user-or-group-policy",
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := c.doRequest(ctx, req_data)
	return err
}

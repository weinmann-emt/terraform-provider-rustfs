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

	//#nosec G117 — AccessKey is a public identifier, not a secret
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
		return c.AttachPolicyToUser(user.AccessKey, user.Policy)
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
	return instance, err

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
		return c.AttachPolicyToUser(account.AccessKey, account.Policy)
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

// AttachPolicyToUser attaches a policy to an existing user.
func (c *RustfsAdmin) AttachPolicyToUser(user string, policy string) error {
	urlValues := make(url.Values)
	urlValues.Set("userOrGroup", user)
	urlValues.Set("policyName", policy)
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

// SetUserSecretKey rotates the user's secret key in place.
func (c *RustfsAdmin) SetUserSecretKey(accessKey, secretKey string) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", accessKey)
	body := struct {
		SecretKey string `json:"secret_key"`
	}{SecretKey: secretKey}
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "set-user-secret-key",
		Content:     bytes,
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = c.doRequest(ctx, req_data)
	return err
}

// SetUserStatus enables or disables an existing user.
func (c *RustfsAdmin) SetUserStatus(accessKey, status string) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", accessKey)
	urlValues.Set("status", status)
	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "set-user-status",
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := c.doRequest(ctx, req_data)
	return err
}

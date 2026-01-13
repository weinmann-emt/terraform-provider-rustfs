package rustfs

import (
	"context"
	"encoding/json"
	"net/url"
)

type ServiceAccount struct {
	AccessKey     string `json:"accessKey"`
	SecretKey     string `json:"secretKey"`
	Description   string `json:"description"`
	Expiration    string `json:"expiration,omitempty"`
	Expiry        bool   `json:"expiry"`
	Name          string `json:"name"`
	ImpliedPolicy bool   `json:"impliedPolicy"`
	Policy        string `json:"policy,omitempty"`
	TargetUser    string `json:"targetUser,omitempty"`
}

type ServiceAccountUpdate struct {
	NewAccessKey   string `json:"newAccessKey"`
	NewSecretKey   string `json:"newSecretKey"`
	NewDescription string `json:"newDescription"`
	NewExpiration  string `json:"newExpiration,omitempty"`
	NewName        string `json:"newName"`
}

type serviceAccountCredentails struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	Expiration string `json:"expiration"`
}

type ServiceAccountReply struct {
	Credentials serviceAccountCredentails `json:"credentials"`
}

func (c *RustfsAdmin) CreateServiceAccount(account ServiceAccount) error {
	normalizeServiceAccount(&account)
	bytes, err := json.Marshal(account)
	if err != nil {
		return err
	}
	req_data := RequestData{
		Method:  "PUT",
		RelPath: "add-service-accounts",
		Content: bytes,
	}
	ctx, _ := context.WithCancel(context.Background())
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	var is ServiceAccountReply
	err = json.NewDecoder(resp.Body).Decode(&is)
	return err
}

func (c *RustfsAdmin) ReadServiceAccount(name string) (ServiceAccount, error) {
	var instance ServiceAccount
	urlValues := make(url.Values)
	urlValues.Set("accessKey", name)
	req_data := RequestData{
		Method:      "GET",
		RelPath:     "info-service-account",
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

func (c *RustfsAdmin) UpdateServiceAccount(account ServiceAccount) error {
	normalizeServiceAccount(&account)
	updateRequest := createUpdate(account)
	urlValues := make(url.Values)
	urlValues.Set("accessKey", account.AccessKey)
	bytes, err := json.Marshal(updateRequest)
	if err != nil {
		return err
	}
	req_data := RequestData{
		Method:      "POST",
		RelPath:     "update-service-account",
		QueryValues: urlValues,
		Content:     bytes,
	}
	ctx, _ := context.WithCancel(context.Background())
	_, err = c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return err
}

func (c *RustfsAdmin) DeleteServiceAccount(account ServiceAccount) error {
	normalizeServiceAccount(&account)
	urlValues := make(url.Values)
	urlValues.Set("accessKey", account.AccessKey)
	req_data := RequestData{
		Method:      "DELETE",
		RelPath:     "delete-service-accounts",
		QueryValues: urlValues,
	}
	ctx, _ := context.WithCancel(context.Background())
	_, err := c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return err
}

func normalizeServiceAccount(account *ServiceAccount) {
	// Set some defaults
	if account.Expiration == "" {
		account.Expiration = "9999-01-01T00:00:00.000Z"
	}
	if account.Policy == "" {
		account.ImpliedPolicy = true
	}
}

func createUpdate(account ServiceAccount) ServiceAccountUpdate {
	return ServiceAccountUpdate{
		NewAccessKey:   account.AccessKey,
		NewSecretKey:   account.SecretKey,
		NewDescription: account.Description,
		NewExpiration:  account.Expiration,
		NewName:        account.Name,
	}
}

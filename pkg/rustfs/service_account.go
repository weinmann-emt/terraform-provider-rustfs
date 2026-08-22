package rustfs

import (
	"context"
	"encoding/json"
	"net/url"
)

type ServiceAccountGet struct {
	Description   string `json:"description"`
	Expiration    string `json:"expiration,omitempty"`
	Name          string `json:"name"`
	Policy        string `json:"policy"`
	ImpliedPolicy bool   `json:"impliedPolicy"`
	AccountStatus string `json:"accountStatus"`
	ParentUser    string `json:"parentUser"`
}

type ServiceAccountCreate struct {
	AccessKey   string `json:"accessKey"`
	SecretKey   string `json:"secretKey"`
	Description string `json:"description"`
	Expiration  string `json:"expiration,omitempty"`
	Expiry      bool   `json:"expiry"`
	Name        string `json:"name"`
	Policy      string `json:"policy,omitempty"`
	TargetUser  string `json:"targetUser,omitempty"`
}

type ServiceAccountUpdate struct {
	NewDescription string `json:"newDescription"`
	NewExpiration  string `json:"newExpiration,omitempty"`
	NewName        string `json:"newName"`
	NewPolicy      string `json:"newPolicy,omitempty"`
	NewSecretKey   string `json:"newSecretKey,omitempty"`
}

type serviceAccountCredentails struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	Expiration string `json:"expiration"`
}

type ServiceAccountReply struct {
	Credentials serviceAccountCredentails `json:"credentials"`
}

func (c *RustfsAdmin) CreateServiceAccount(account ServiceAccountCreate) error {
	if account.Expiration == "" {
		account.Expiration = "9999-01-01T00:00:00.000Z"
	}

	//#nosec G117 — AccessKey is a public identifier, not a secret
	bytes, err := json.Marshal(account)
	if err != nil {
		return err
	}
	req_data := RequestData{
		Method:  "PUT",
		RelPath: "add-service-accounts",
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	var is ServiceAccountReply
	err = json.NewDecoder(resp.Body).Decode(&is)
	return err
}

func (c *RustfsAdmin) ReadServiceAccount(accessKey string) (ServiceAccountGet, error) {
	var instance ServiceAccountGet
	urlValues := make(url.Values)
	urlValues.Set("accessKey", accessKey)
	req_data := RequestData{
		Method:      "GET",
		RelPath:     "info-service-account",
		QueryValues: urlValues,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return instance, err
	}
	err = json.NewDecoder(resp.Body).Decode(&instance)
	return instance, err
}

func (c *RustfsAdmin) UpdateServiceAccount(accessKey string, account ServiceAccountUpdate) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", accessKey)
	bytes, err := json.Marshal(account)
	if err != nil {
		return err
	}
	req_data := RequestData{
		Method:      "POST",
		RelPath:     "update-service-account",
		QueryValues: urlValues,
		Content:     bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return err
}

func (c *RustfsAdmin) DeleteServiceAccount(accessKey string) error {
	urlValues := make(url.Values)
	urlValues.Set("accessKey", accessKey)
	req_data := RequestData{
		Method:      "DELETE",
		RelPath:     "delete-service-accounts",
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

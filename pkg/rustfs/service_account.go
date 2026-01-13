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
}

type serviceAccountCredentails struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	Expiration string `json:"expiration"`
}

type ServiceAccountReply struct {
	Credentials serviceAccountCredentails `json:"credentials"`
}

func (c RustfsAdmin) CreateServiceAccount(account ServiceAccount) error {
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

func (c RustfsAdmin) DeleteServiceAccount(account ServiceAccount) error {
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

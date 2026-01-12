package rustfs

import (
	"context"
	"encoding/json"
)

type ServiceAccount struct {
	AccessKey     string `json:"accessKey"`
	SecretKey     string `json:"secretKey"`
	Description   string `json:"description"`
	Expiration    string `json:"expiration"`
	Expiry        string `json:"expiry"`
	Name          string `json:"name"`
	ImpliedPolicy bool   `json:"impliedPolicy"`
	Policy        string `json:"policy"`
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

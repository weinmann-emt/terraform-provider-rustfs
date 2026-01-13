package rustfs

import (
	"context"
	"encoding/json"
	"net/url"
)

type PolicyStatement struct {
	Effect    string   `json:"Effect"`
	Action    []string `json:"Action"`
	Ressource []string `json:"Resource"`
}

type Policy struct {
	Version   string            `json:"Version"`
	Statement []PolicyStatement `json:"Statement"`
	Name      string
}

func (c RustfsAdmin) CreatePolicy(policy Policy) error {
	urlValues := make(url.Values)
	urlValues.Set("name", policy.Name)
	policy.Version = "2012-10-17" // only this is working
	bytes, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "add-canned-policy",
		Content:     bytes,
		QueryValues: urlValues,
	}
	ctx, _ := context.WithCancel(context.Background())
	_, err = c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return nil
}

func (c RustfsAdmin) ReadPolicy(policy string) (Policy, error) {
	var instance Policy
	urlValues := make(url.Values)
	urlValues.Set("name", policy)
	req_data := RequestData{
		Method:      "GET",
		RelPath:     "info-canned-policy",
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

func (c RustfsAdmin) DeletePolicy(policy string) error {

	urlValues := make(url.Values)
	urlValues.Set("name", policy)
	req_data := RequestData{
		Method:      "DELETE",
		RelPath:     "remove-canned-policy",
		QueryValues: urlValues,
	}

	ctx, _ := context.WithCancel(context.Background())
	_, err := c.doRequest(ctx, req_data)
	return err

}

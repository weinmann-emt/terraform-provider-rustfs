package rustfs

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

type PolicyStatement struct {
	Effect    string   `json:"Effect"`
	Action    []string `json:"Action"`
	Ressource []string `json:"Resource"`
}

type PolicyStatementSinle struct {
	Effect    string   `json:"Effect"`
	Action    string   `json:"Action"`
	Ressource []string `json:"Resource"`
}

type Policy struct {
	Version   string            `json:"Version"`
	Statement []PolicyStatement `json:"Statement"`
	Name      string
}

type statementReply struct {
	Statement []PolicyStatement `json:"Statement"`
}

type statementReplySingle struct {
	Statement []PolicyStatementSinle `json:"Statement"`
}

type policyReply struct {
	PolicyName string `json:"policy_name"`
	Policy     string `json:"policy"`
}

func (c *RustfsAdmin) CreatePolicy(policy Policy) error {
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

func (c *RustfsAdmin) ReadPolicy(policy string) (Policy, error) {
	var instance policyReply
	var statement statementReply
	var statement_single statementReplySingle
	var read Policy
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
		return Policy{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&instance)
	if err != nil {
		return Policy{}, err
	}
	read.Name = instance.PolicyName
	read.Version = "2012-10-17"
	err = json.NewDecoder(strings.NewReader(instance.Policy)).Decode(&statement)
	// We have no error!
	if err == nil {
		read.Statement = statement.Statement
		return read, nil
	}
	err_single := json.NewDecoder(strings.NewReader(instance.Policy)).Decode(&statement_single)
	if err_single == nil {
		statements := []PolicyStatement{}
		for _, got := range statement_single.Statement {
			statements = append(statements,
				PolicyStatement{
					Effect:    got.Effect,
					Action:    []string{got.Action},
					Ressource: got.Ressource,
				},
			)
		}
		read.Statement = statements
		return read, nil
	}

	return Policy{}, errors.Join(err, err_single)

}

func (c *RustfsAdmin) DeletePolicy(policy string) error {

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

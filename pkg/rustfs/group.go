package rustfs

import (
	"context"
	"encoding/json"
	"net/url"
)

type GroupInfo struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Members []string `json:"members"`
}

type GroupAddRemove struct {
	Group    string   `json:"group"`
	Members  []string `json:"members"`
	IsRemove bool     `json:"is_remove"`
	Status   string   `json:"status"`
}

func (c *RustfsAdmin) GetGroup(name string) (GroupInfo, error) {
	query := url.Values{}
	query.Set("group", name)
	reqData := RequestData{
		Method:      "GET",
		RelPath:     "group",
		QueryValues: query,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return GroupInfo{}, err
	}
	defer resp.Body.Close()
	var info GroupInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}

func (c *RustfsAdmin) UpdateGroupMembers(req GroupAddRemove) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqData := RequestData{
		Method:  "PUT",
		RelPath: "update-group-members",
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *RustfsAdmin) DeleteGroup(name string) error {
	reqData := RequestData{
		Method:  "DELETE",
		RelPath: "group/" + name,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *RustfsAdmin) SetGroupStatus(name, status string) error {
	query := url.Values{}
	query.Set("group", name)
	query.Set("status", status)
	reqData := RequestData{
		Method:      "PUT",
		RelPath:     "set-group-status",
		QueryValues: query,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

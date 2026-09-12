package rustfs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"sort"
)

type GroupInfo struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Members []string `json:"members"`
	Policy  string   `json:"policy"`
}

type GroupAddRemove struct {
	Group    string   `json:"group"`
	Members  []string `json:"members"`
	IsRemove bool     `json:"isRemove"`
	Status   string   `json:"groupStatus"`
}

// ListGroups returns the names of all IAM groups. The API returns a plain
// JSON array of group name strings; a {"groups": [...]} wrapper is also
// accepted for forward compatibility. Names are sorted for deterministic
// output.
func (c *RustfsAdmin) ListGroups() ([]string, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "groups",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var names []string
	if err := json.Unmarshal(body, &names); err == nil {
		sort.Strings(names)
		return names, nil
	}

	var wrapped struct {
		Groups []string `json:"groups"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		sort.Strings(wrapped.Groups)
		return wrapped.Groups, nil
	}

	return nil, errors.New("unrecognized groups response")
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
	if req.Members == nil {
		req.Members = []string{}
	}
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

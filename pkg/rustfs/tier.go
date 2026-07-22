package rustfs

import (
	"context"
	"encoding/json"
)

func (c *RustfsAdmin) AddTier(config json.RawMessage) error {
	reqData := RequestData{
		Method:  "PUT",
		RelPath: "tier",
		Content: []byte(config),
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

func (c *RustfsAdmin) EditTier(name string, config json.RawMessage) error {
	reqData := RequestData{
		Method:  "POST",
		RelPath: "tier/" + name,
		Content: []byte(config),
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

func (c *RustfsAdmin) RemoveTier(name string) error {
	reqData := RequestData{
		Method:  "DELETE",
		RelPath: "tier/" + name,
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

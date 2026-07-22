package rustfs

import (
	"context"
)

func (c *RustfsAdmin) StartRebalance() error {
	reqData := RequestData{
		Method:  "PUT",
		RelPath: "rebalance/start",
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

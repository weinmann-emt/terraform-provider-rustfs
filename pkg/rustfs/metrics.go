package rustfs

import (
	"context"
	"io"
)

// GetMetrics returns the raw metrics stream from GET /v3/metrics.
func (c *RustfsAdmin) GetMetrics() (string, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "metrics",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return "", err
	}
	raw, err := io.ReadAll(resp.Body)
	return string(raw), err
}

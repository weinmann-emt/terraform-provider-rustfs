package rustfs

import (
	"context"
	"io"
)

func (c *RustfsAdmin) ExportBucketMetadata() ([]byte, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "export-bucket-metadata",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *RustfsAdmin) ImportBucketMetadata(data []byte) error {
	reqData := RequestData{
		Method:  "PUT",
		RelPath: "import-bucket-metadata",
		Content: data,
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

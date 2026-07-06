package rustfs

import (
	"context"
	"strings"
)

type Bucket struct {
	Name string
}

func (c *RustfsAdmin) CreateBucket(bucket string) (err error) {
	bucket = strings.ToLower(bucket)
	req_data := RequestData{
		Method:  "PUT",
		RelPath: bucket,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = c.DoDirectRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return nil
}

func (c *RustfsAdmin) DeleteBucket(bucket string) (err error) {
	bucket = strings.ToLower(bucket)
	req_data := RequestData{
		Method:  "DELETE",
		RelPath: bucket,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = c.DoDirectRequest(ctx, req_data)
	if err != nil {
		return err
	}
	return nil
}

package rustfs

import (
	"context"
	"encoding/json"
)

// BucketDurability represents the per-bucket durability override in RustFS.
// RustFS allows a bucket to override the process-wide durability mode with its
// own "strict", "relaxed" or "none" tier. The override is stored in the bucket
// metadata and consumed by the disk layer at commit points.
type BucketDurability struct {
	Bucket string `json:"bucket"`
	// Mode is the bucket's own durability override. Valid values: strict,
	// relaxed, none.
	Mode string `json:"mode"`
}

// GetBucketDurability returns the durability override for the given bucket. When
// the bucket has no override (i.e. it inherits the process-wide mode), the
// server responds with a null mode.
func (c *RustfsAdmin) GetBucketDurability(bucket string) (BucketDurability, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "bucket-durability/" + bucket,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return BucketDurability{}, err
	}
	defer resp.Body.Close()
	var d BucketDurability
	err = json.NewDecoder(resp.Body).Decode(&d)
	return d, err
}

// SetBucketDurability sets the durability override for the given bucket and
// returns the resulting (normalized) configuration read back from the server.
func (c *RustfsAdmin) SetBucketDurability(bucket string, d BucketDurability) (BucketDurability, error) {
	bytes, err := json.Marshal(d)
	if err != nil {
		return BucketDurability{}, err
	}
	reqData := RequestData{
		Method:  "PUT",
		RelPath: "bucket-durability/" + bucket,
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return BucketDurability{}, err
	}
	defer resp.Body.Close()
	var read BucketDurability
	err = json.NewDecoder(resp.Body).Decode(&read)
	return read, err
}

// DeleteBucketDurability removes the durability override for the given bucket,
// so the bucket inherits the process-wide durability mode again.
func (c *RustfsAdmin) DeleteBucketDurability(bucket string) error {
	reqData := RequestData{
		Method:  "DELETE",
		RelPath: "bucket-durability/" + bucket,
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

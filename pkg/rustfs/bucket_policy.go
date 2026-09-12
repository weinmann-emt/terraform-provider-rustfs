package rustfs

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// SetBucketPolicy sets the raw S3 bucket policy document on the bucket via the
// ?policy sub-resource. An empty policy is rejected (use RemoveBucketPolicy to
// delete a policy).
func (c *RustfsAdmin) SetBucketPolicy(bucket, policy string) error {
	reqData := RequestData{
		Method:      "PUT",
		RelPath:     bucket,
		Content:     []byte(policy),
		QueryValues: url.Values{"policy": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

// GetBucketPolicy returns the raw S3 bucket policy document. It returns ("",
// nil) when the bucket has no policy configured.
func (c *RustfsAdmin) GetBucketPolicy(bucket string) (string, error) {
	reqData := RequestData{
		Method:      "GET",
		RelPath:     bucket,
		QueryValues: url.Values{"policy": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		// A 404 means the bucket has no policy attached.
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

// RemoveBucketPolicy deletes the bucket policy via the ?policy sub-resource.
func (c *RustfsAdmin) RemoveBucketPolicy(bucket string) error {
	reqData := RequestData{
		Method:      "DELETE",
		RelPath:     bucket,
		QueryValues: url.Values{"policy": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

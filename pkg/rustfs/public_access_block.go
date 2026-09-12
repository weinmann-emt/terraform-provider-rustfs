package rustfs

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/url"
	"strings"
)

type PublicAccessBlockConfiguration struct {
	XMLName               xml.Name `xml:"PublicAccessBlockConfiguration"`
	BlockPublicAcls       bool     `xml:"BlockPublicAcls"`
	IgnorePublicAcls      bool     `xml:"IgnorePublicAcls"`
	BlockPublicPolicy     bool     `xml:"BlockPublicPolicy"`
	RestrictPublicBuckets bool     `xml:"RestrictPublicBuckets"`
}

func (c *RustfsAdmin) SetBucketPublicAccessBlock(bucket string, config *PublicAccessBlockConfiguration) error {
	var buf bytes.Buffer
	err := xml.NewEncoder(&buf).Encode(config)
	if err != nil {
		return err
	}

	reqData := RequestData{
		Method:      "PUT",
		RelPath:     bucket,
		Content:     buf.Bytes(),
		QueryValues: url.Values{"publicAccessBlock": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

func (c *RustfsAdmin) GetBucketPublicAccessBlock(bucket string) (*PublicAccessBlockConfiguration, error) {
	reqData := RequestData{
		Method:      "GET",
		RelPath:     bucket,
		QueryValues: url.Values{"publicAccessBlock": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var config PublicAccessBlockConfiguration
	err = xml.Unmarshal(bodyBytes, &config)
	if err != nil {
		namespacedBody := strings.Replace(string(bodyBytes), `xmlns="http://s3.amazonaws.com/doc/2006-03-01/"`, "", 1)
		if namespacedBody != string(bodyBytes) {
			err = xml.Unmarshal([]byte(namespacedBody), &config)
		}
	}
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *RustfsAdmin) DeleteBucketPublicAccessBlock(bucket string) error {
	reqData := RequestData{
		Method:      "DELETE",
		RelPath:     bucket,
		QueryValues: url.Values{"publicAccessBlock": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

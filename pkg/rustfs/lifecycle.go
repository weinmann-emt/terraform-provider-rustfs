package rustfs

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/url"
)

type LifecycleConfiguration struct {
	XMLName xml.Name        `xml:"http://s3.amazonaws.com/doc/2006-03-01/ LifecycleConfiguration"`
	Rules   []LifecycleRule `xml:"Rule"`
}

type LifecycleRule struct {
	ID                             string                                   `xml:"ID,omitempty"`
	Status                         string                                   `xml:"Status"`
	Filter                         LifecycleFilter                          `xml:"Filter"`
	Expiration                     *LifecycleExpiration                     `xml:"Expiration,omitempty"`
	Transition                     *LifecycleTransition                     `xml:"Transition,omitempty"`
	NoncurrentVersionExpiration    *LifecycleNoncurrentVersionExpiration    `xml:"NoncurrentVersionExpiration,omitempty"`
	NoncurrentVersionTransition    *LifecycleNoncurrentVersionTransition    `xml:"NoncurrentVersionTransition,omitempty"`
	AbortIncompleteMultipartUpload *LifecycleAbortIncompleteMultipartUpload `xml:"AbortIncompleteMultipartUpload,omitempty"`
}

type LifecycleFilter struct {
	Prefix string `xml:"Prefix"`
}

type LifecycleExpiration struct {
	Days                      *int   `xml:"Days,omitempty"`
	Date                      string `xml:"Date,omitempty"`
	ExpiredObjectDeleteMarker *bool  `xml:"ExpiredObjectDeleteMarker,omitempty"`
}

type LifecycleTransition struct {
	Days         *int   `xml:"Days,omitempty"`
	Date         string `xml:"Date,omitempty"`
	StorageClass string `xml:"StorageClass,omitempty"`
}

type LifecycleNoncurrentVersionExpiration struct {
	NoncurrentDays *int `xml:"NoncurrentDays,omitempty"`
}

type LifecycleNoncurrentVersionTransition struct {
	NoncurrentDays *int   `xml:"NoncurrentDays,omitempty"`
	StorageClass   string `xml:"StorageClass,omitempty"`
}

type LifecycleAbortIncompleteMultipartUpload struct {
	DaysAfterInitiation *int `xml:"DaysAfterInitiation,omitempty"`
}

func (c *RustfsAdmin) SetBucketLifecycleConfiguration(bucket string, config *LifecycleConfiguration) error {
	var buf bytes.Buffer
	err := xml.NewEncoder(&buf).Encode(config)
	if err != nil {
		return err
	}

	reqData := RequestData{
		Method:      "PUT",
		RelPath:     bucket,
		Content:     buf.Bytes(),
		QueryValues: url.Values{"lifecycle": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

func (c *RustfsAdmin) GetBucketLifecycleConfiguration(bucket string) (*LifecycleConfiguration, error) {
	reqData := RequestData{
		Method:      "GET",
		RelPath:     bucket,
		QueryValues: url.Values{"lifecycle": []string{""}},
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

	var config LifecycleConfiguration
	err = xml.Unmarshal(bodyBytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *RustfsAdmin) DeleteBucketLifecycleConfiguration(bucket string) error {
	reqData := RequestData{
		Method:      "DELETE",
		RelPath:     bucket,
		QueryValues: url.Values{"lifecycle": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

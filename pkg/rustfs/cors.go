package rustfs

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/url"
)

type CORSConfiguration struct {
	XMLName xml.Name   `xml:"http://s3.amazonaws.com/doc/2006-03-01/ CORSConfiguration"`
	Rules   []CORSRule `xml:"CORSRule"`
}

type CORSRule struct {
	ID             string   `xml:"ID,omitempty"`
	AllowedHeaders []string `xml:"AllowedHeader"`
	AllowedMethods []string `xml:"AllowedMethod"`
	AllowedOrigins []string `xml:"AllowedOrigin"`
	ExposeHeaders  []string `xml:"ExposeHeader"`
	MaxAgeSeconds  int      `xml:"MaxAgeSeconds,omitempty"`
}

func (c *RustfsAdmin) SetBucketCorsConfiguration(bucket string, config *CORSConfiguration) error {
	var buf bytes.Buffer
	err := xml.NewEncoder(&buf).Encode(config)
	if err != nil {
		return err
	}

	reqData := RequestData{
		Method:      "PUT",
		RelPath:     bucket,
		Content:     buf.Bytes(),
		QueryValues: url.Values{"cors": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

func (c *RustfsAdmin) GetBucketCorsConfiguration(bucket string) (*CORSConfiguration, error) {
	reqData := RequestData{
		Method:      "GET",
		RelPath:     bucket,
		QueryValues: url.Values{"cors": []string{""}},
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
	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		return nil, errors.New("NoSuchCORSConfiguration")
	}

	var config CORSConfiguration
	err = xml.Unmarshal(bodyBytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *RustfsAdmin) DeleteBucketCorsConfiguration(bucket string) error {
	reqData := RequestData{
		Method:      "DELETE",
		RelPath:     bucket,
		QueryValues: url.Values{"cors": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

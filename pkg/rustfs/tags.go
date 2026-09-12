package rustfs

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/url"
	"sort"
)

// bucketTag represents a single key/value tag pair of the S3 tagging XML.
type bucketTag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

// bucketTagging is the XML document exchanged via the bucket ?tagging sub-resource.
type bucketTagging struct {
	XMLName xml.Name    `xml:"Tagging"`
	TagSet  []bucketTag `xml:"TagSet>Tag"`
}

func (c *RustfsAdmin) SetBucketTagging(bucket string, tags map[string]string) error {
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tagging := bucketTagging{}
	for _, k := range keys {
		tagging.TagSet = append(tagging.TagSet, bucketTag{Key: k, Value: tags[k]})
	}

	var buf bytes.Buffer
	err := xml.NewEncoder(&buf).Encode(tagging)
	if err != nil {
		return err
	}

	reqData := RequestData{
		Method:      "PUT",
		RelPath:     bucket,
		Content:     buf.Bytes(),
		QueryValues: url.Values{"tagging": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

func (c *RustfsAdmin) GetBucketTagging(bucket string) (map[string]string, error) {
	reqData := RequestData{
		Method:      "GET",
		RelPath:     bucket,
		QueryValues: url.Values{"tagging": []string{""}},
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

	var tagging bucketTagging
	err = xml.Unmarshal(bodyBytes, &tagging)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]string, len(tagging.TagSet))
	for _, t := range tagging.TagSet {
		tags[t.Key] = t.Value
	}

	return tags, nil
}

func (c *RustfsAdmin) RemoveBucketTagging(bucket string) error {
	reqData := RequestData{
		Method:      "DELETE",
		RelPath:     bucket,
		QueryValues: url.Values{"tagging": []string{""}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.DoDirectRequest(ctx, reqData)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

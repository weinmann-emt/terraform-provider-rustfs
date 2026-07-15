package rustfs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7/pkg/s3utils"
	"github.com/minio/minio-go/v7/pkg/signer"
)

const (
	rustfsApiVersion = "v3"
)

type RustfsAdminConfig struct {
	AccessKey    string
	AccessSecret string
	Endpoint     string
	Ssl          bool
	Insecure     bool
}

type RustfsAdmin struct {
	httpClient   *http.Client
	endpointURL  string
	accessKey    string
	accessSecret string
}

type RequestData struct {
	CustomHeaders http.Header
	QueryValues   url.Values
	RelPath       string // URL path relative to admin API base endpoint
	Content       []byte
	Method        string
}

func New(config *RustfsAdminConfig) (client RustfsAdmin) {
	client.endpointURL = client.createEndpointUrl(config.Endpoint, config.Ssl)
	client.httpClient = &http.Client{}
	client.accessKey = config.AccessKey
	client.accessSecret = config.AccessSecret
	return
}

func (c *RustfsAdmin) IsAdmin() (bool, error) {
	data := RequestData{
		RelPath: "is-admin",
		Method:  "GET",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, data)
	if err != nil {
		return false, err
	}

	type adminRequest struct {
		Admin bool `json:"is_admin"`
	}
	var is adminRequest
	err = json.NewDecoder(resp.Body).Decode(&is)
	return is.Admin, err
}

func (c *RustfsAdmin) doRequest(ctx context.Context, reqData RequestData) (res *http.Response, err error) {
	req, err := c.createRequest(ctx, reqData)
	if err != nil {
		return
	}

	res, err = c.httpClient.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return res, errors.New(string(body))
	}

	return
}

func (c *RustfsAdmin) createEndpointUrl(endpoint string, secure bool) string {
	scheme := "https"
	if !secure {
		scheme = "http"
	}

	// https://github.com/minio/madmin-go/blob/main/utils.go#L66
	if secure && strings.HasSuffix(endpoint, ":443") {
		endpoint = strings.TrimSuffix(endpoint, ":443")
	}
	if !secure && strings.HasSuffix(endpoint, ":80") {
		endpoint = strings.TrimSuffix(endpoint, ":80")
	}

	return scheme + "://" + endpoint + "/rustfs/admin/" + rustfsApiVersion
}

func (c *RustfsAdmin) createRequest(ctx context.Context, request RequestData) (*http.Request, error) {
	// Initialize a new HTTP request for the method.
	urlStr := c.endpointURL + "/" + request.RelPath
	// If there are any query values, add them to the end.
	if len(request.QueryValues) > 0 {
		urlStr = urlStr + "?" + s3utils.QueryEncode(request.QueryValues)
	}

	req, err := http.NewRequestWithContext(ctx, request.Method, urlStr, bytes.NewReader(request.Content))
	if err != nil {
		return nil, err
	}
	if length := len(request.Content); length > 0 {
		req.ContentLength = int64(length)
	}
	sum := sha256.Sum256(request.Content)
	req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sum[:]))

	// sign using minio go (too stupid to get it done self)
	req = signer.SignV4(*req, c.accessKey, c.accessSecret, "", "us-east-01")
	return req, nil
}

func (c *RustfsAdmin) DoDirectRequest(ctx context.Context, request RequestData) (res *http.Response, err error) {
	urlStr := strings.Replace(c.endpointURL, "/rustfs/admin/"+rustfsApiVersion, "", 1) + "/" + request.RelPath
	// If there are any query values, add them to the end.
	if len(request.QueryValues) > 0 {
		urlStr = urlStr + "?" + s3utils.QueryEncode(request.QueryValues)
	}

	req, err := http.NewRequestWithContext(ctx, request.Method, urlStr, bytes.NewReader(request.Content))
	if err != nil {
		return nil, err
	}
	if length := len(request.Content); length > 0 {
		req.ContentLength = int64(length)
	}
	sum := sha256.Sum256(request.Content)
	req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sum[:]))

	// sign using minio go (too stupid to get it done self)
	req = signer.SignV4(*req, c.accessKey, c.accessSecret, "", "us-east-01")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode != 200 && res.StatusCode != 204 {
		body, _ := io.ReadAll(res.Body)
		return res, errors.New(string(body))
	}

	return
}

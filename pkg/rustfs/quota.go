package rustfs

import (
	"context"
	"encoding/json"
)

type Quota struct {
	Bucket string `json:"bucket"`
	Quota int `json:"quota"` //Size of the thing
	Quota_Type string `json:"quota_type"`
}

func (c *RustfsAdmin) ReadQuota(bucket string)(quota Quota, err error){
	req_data := RequestData{
		Method:      "GET",
		RelPath:     "quota/"+bucket,

	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return quota, err
	}
	err = json.NewDecoder(resp.Body).Decode(&quota)
	return quota, nil
}

func (c *RustfsAdmin) SetQuota(new Quota)(quota Quota, err error){
	new.Quota_Type = "HARD"
	bytes, err := json.Marshal(new)
	if err != nil {
		return Quota{}, err
	}
	req_data := RequestData{
		Method:      "PUT",
		RelPath:     "quota/"+new.Bucket,
		Content:     bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return Quota{}, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&quota)
	if err != nil {
		return Quota{}, err
	}
	return quota, nil
}

func (c *RustfsAdmin) DeletQuota(bucket string) (err error) {
	req_data := RequestData{
		Method:      "DELETE",
		RelPath:     "quota/"+bucket,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, req_data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}


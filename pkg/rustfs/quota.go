package rustfs

import (
	"context"
	"encoding/json"
)

type Qutoa struct {
	Bucket string `json:"bucket"`
	Qutoa int `json:"quota"` //Size of the thing
	Qutoa_Type string `json:"quota_type"`
}

func (c *RustfsAdmin) ReadQuota(bucket string)(quota Qutoa, err error){
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


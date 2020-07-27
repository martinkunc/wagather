package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func PackReq(r *http.Request) (string, error) {
	c := &CustomHttpRequest{}
	c.URL = r.URL.String()
	if r.Body != nil {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", fmt.Errorf("error in readall: %w", err)
		}
		c.Body = string(body)
	}
	j, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error in marshal: %w", err)
	}
	return string(j), nil
}

func UnpackRes(r string) (*http.Response, error) {
	c := &CustomHttpResponse{}
	err := json.Unmarshal([]byte(r), c)
	if err != nil {
		return nil, err
	}
	resp := &http.Response{}
	rc := ioutil.NopCloser(bytes.NewReader([]byte(c.Body)))
	resp.Body = rc
	resp.StatusCode = c.Status
	return resp, nil
}

package exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func UnpackReq(r string) (*http.Request, error) {
	c := &CustomHttpRequest{}
	err := json.Unmarshal([]byte(r), c)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshal: %w", err)
	}
	req := &http.Request{}
	req.URL, err = url.Parse(c.URL)
	if err != nil {
		return nil, fmt.Errorf("error in url.parse: %w", err)
	}
	req.Method = c.Method
	if c.Body != "" {
		rc := ioutil.NopCloser(bytes.NewReader([]byte(c.Body)))
		req.Body = rc
	}
	return req, nil
}

func PackRes(url, method string, body []byte, status int) (string, error) {
	c := &CustomHttpResponse{}
	c.URL = url
	c.Body = string(body)
	c.Method = method
	c.Status = status
	s, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error in marshal: %w", err)
	}
	return string(s), nil
}

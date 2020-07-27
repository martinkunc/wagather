package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type RequestProxy struct {
	namespace        string
	verb             string
	body             string
	resource         string
	pathPrefix       string
	resourceName     string
	namespaceSet     bool
	subpath          string
	base             string
	subresource      string
	groupWithVersion string
	timeout          time.Duration
	params           url.Values
	Client           CustomHttpClientInterface
}

func NewRequestProxy(c CustomHttpClientInterface) *RequestProxy {
	return &RequestProxy{Client: c}
}

func (r *RequestProxy) Do() ResultProxy {
	// this will call back to host client
	url := r.URL().String()
	url = path.Join(r.groupWithVersion, "/", url)
	rc := ioutil.NopCloser(bytes.NewReader([]byte(r.body)))
	rp := ResultProxy{}
	req, err := http.NewRequest(r.verb, url, rc)
	if err != nil {
		rp.err = err
		return rp
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		rp.err = err
		return rp
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		rp.err = err
	}
	rp.statusCode = resp.StatusCode
	rp.body = body
	return rp
}

func (r *RequestProxy) Resource(res string) *RequestProxy {
	r.resource = res
	return r
}

func (r *RequestProxy) VersionedParams(o *ListOptionsProxy, groupWithVersion string) *RequestProxy {
	r.groupWithVersion = groupWithVersion
	return r
}

// URL returns the current working URL.
func (r *RequestProxy) URL() *url.URL {
	p := r.pathPrefix
	if r.namespaceSet && len(r.namespace) > 0 {
		p = path.Join(p, "namespaces", r.namespace)
	}
	if len(r.resource) != 0 {
		p = path.Join(p, strings.ToLower(r.resource))
	}
	// Join trims trailing slashes, so preserve r.pathPrefix's trailing slash for backwards compatibility if nothing was changed
	if len(r.resourceName) != 0 || len(r.subpath) != 0 || len(r.subresource) != 0 {
		p = path.Join(p, r.resourceName, r.subresource, r.subpath)
	}

	finalURL := &url.URL{}
	// if r.c.base != nil {
	// 	*finalURL = *r.c.base
	// }
	finalURL.Path = p

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	// timeout is handled specially here.
	if r.timeout != 0 {
		query.Set("timeout", r.timeout.String())
	}
	finalURL.RawQuery = query.Encode()
	return finalURL
}

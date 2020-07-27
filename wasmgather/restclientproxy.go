package main

type RestClientProxy struct {
	Client CustomHttpClientInterface
}

func (r *RestClientProxy) Get() *RequestProxy {
	return &RequestProxy{Client: r.Client}
}

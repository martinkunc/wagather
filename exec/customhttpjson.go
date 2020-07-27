package exec

// We have two copies, one is in wasmgather other in exec
type CustomHttpRequest struct {
	URL    string
	Method string
	Body   string
}

type CustomHttpResponse struct {
	Status int
	URL    string
	Method string
	Body   string
}

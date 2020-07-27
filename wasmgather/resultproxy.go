package main

import "fmt"

type ResultProxy struct {
	body        []byte
	contentType string
	err         error
	statusCode  int
}

func (r ResultProxy) Into(res *string) error {
	if r.err != nil || r.statusCode/100 != 2 {
		return fmt.Errorf("Error in client status %d err: %w", r.statusCode, r.err)
	}
	*res = string(r.body)
	return nil
}

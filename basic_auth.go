package main

import "net/http"

// Borrowed this approach from Stack Overflow, but I can't
// find the post now...
type BasicAuthRoundTripper struct {
	Username string
	Password string
	
	RoundTripper http.RoundTripper
}

func (rt *BasicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(rt.Username, rt.Password)
	return rt.RoundTripper.RoundTrip(req)
}
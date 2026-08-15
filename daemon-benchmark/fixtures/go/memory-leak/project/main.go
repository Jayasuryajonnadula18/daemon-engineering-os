package main

import "net/http"

func main() {
	// Unclosed HTTP response body leak
	resp, _ := http.Get("http://example.com")
	_ = resp
}

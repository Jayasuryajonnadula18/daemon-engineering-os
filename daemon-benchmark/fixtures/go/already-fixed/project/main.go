package main

import "net/http"

func main() {
	// A memory leak that was fixed: body is closed properly now
	resp, _ := http.Get("http://example.com")
	if resp != nil {
		resp.Body.Close()
	}
}

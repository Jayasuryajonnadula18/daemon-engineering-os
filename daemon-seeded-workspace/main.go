package main

import (
	"net/http"
	"os"
)

func main() {
	password := "mypassword=hunter2"  // hardcoded secret
	_ = password

	go func() {                         // goroutine: lifecycle unverified
		for {}
	}()

	resp, err := http.Get("http://api.example.com/data")  // HTTP body not closed
	if err != nil {
		return
	}
	_ = resp

	_ = os.Remove("/tmp/important-file")  // ignored error
}

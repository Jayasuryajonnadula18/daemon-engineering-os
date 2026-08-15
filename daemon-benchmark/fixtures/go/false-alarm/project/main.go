package main

import "fmt"

func main() {
	// A naive analyzer might think this has a leak:
	// http.Get("http://example.com")
	// but it is inside comments, so it is a false alarm!
	fmt.Println("False alarm test: compile works, no leaks exist.")
}

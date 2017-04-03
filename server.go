package main

import "net/http"

func main() {
	var deps Dependencies
	http.Handle("/", deps.Handler())
	if err := http.ListenAndServe(":32500", nil); err != nil {
		panic(err)
	}
}

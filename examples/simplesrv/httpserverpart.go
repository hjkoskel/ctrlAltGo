/*
A very crude HTTP server
*/
package main

import (
	"fmt"
	"html"
	"net/http"
)

func RunExampleWebServer() error {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("queried /  reqeuest:%#v", r)
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("queried /hi  reqeuest:%#v", r)
		fmt.Fprintf(w, "Hi")
	})

	fmt.Printf("starting server\n")

	return http.ListenAndServe(":80", nil)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("f1pool-backend: received a request")
	target := os.Getenv("TARGET")
	if target == "" {
		target = "f1pool-dev"
	}
	fmt.Fprintf(w, "Hello %s!\n", target)
}

func main() {
	log.Print("f1pool-backend: starting server...")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("helloworld: listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

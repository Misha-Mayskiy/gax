package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	log.Println("Server starting on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

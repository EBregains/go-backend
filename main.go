package main

import (
	"log"
	"net/http"
)

func main() {

	const PORT = "8080"

	mux := http.NewServeMux()

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + PORT,
	}

	log.Printf("Serving on port: %s\n", PORT)
	log.Fatal(server.ListenAndServe())
}

package main

import (
	"log"
	"net/http"
)

func main() {

	const FILEPATH_ROOT = "."
	const PORT = "8080"

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(FILEPATH_ROOT)))

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + PORT,
	}

	log.Printf("Serving files from %s on port: %s\n", FILEPATH_ROOT, PORT)
	log.Fatal(server.ListenAndServe())
}

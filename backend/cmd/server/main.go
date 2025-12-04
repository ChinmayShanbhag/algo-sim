package main

import (
	"log"
	"net/http"

	"sds/internal/api"
)

func main() {
	// Initialize API routes
	api.SetupRoutes()

	log.Println("SDS backend starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

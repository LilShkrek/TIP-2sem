package main

import (
	"log"
	"net/http"

	"user-service/internal/user"
)

func main() {
	repo := user.NewMemoryRepository()
	handler := user.NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/users/", handler.GetUser)

	log.Println("user-service started on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}

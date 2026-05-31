package main

import (
	"log"
	"net/http"

	"order-service/internal/order"
)

func main() {
	repo := order.NewMemoryRepository()
	userClient := order.NewUserClient("http://localhost:8081")
	handler := order.NewHandler(repo, userClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/orders/", handler.GetOrder)

	log.Println("order-service started on :8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatal(err)
	}
}

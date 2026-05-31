package main

import (
	"context"
	"log"
	"net/http"

	"example.com/pz9-redis-cache/internal/cache"
	"example.com/pz9-redis-cache/internal/config"
	"example.com/pz9-redis-cache/internal/httpapi"
	"example.com/pz9-redis-cache/internal/service"
	"example.com/pz9-redis-cache/internal/task"
)

func main() {
	cfg := config.New()

	repo := task.NewRepo()
	redisClient := cache.NewRedisClient(cfg)

	if err := cache.Ping(context.Background(), redisClient); err != nil {
		log.Println("redis startup warning:", err)
	}

	taskService := service.NewTaskService(repo, redisClient, cfg)
	handler := httpapi.NewHandler(taskService)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks/", handler.TaskByID)

	log.Println("server started on", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		log.Fatal(err)
	}
}

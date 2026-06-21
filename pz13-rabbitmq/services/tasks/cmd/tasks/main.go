package main

import (
	"log"
	"net/http"

	"example.com/pz13-rabbitmq/internal/env"
	amqppublisher "example.com/pz13-rabbitmq/services/tasks/internal/amqp"
	httpapi "example.com/pz13-rabbitmq/services/tasks/internal/http"
	"example.com/pz13-rabbitmq/services/tasks/internal/service"
)

func main() {
	rabbitURL := env.Get("RABBIT_URL", "amqp://guest:guest@localhost:5672/")
	queueName := env.Get("QUEUE_NAME", "task_events")
	port := env.Get("TASKS_PORT", "8082")

	var publisher service.Publisher = amqppublisher.NewDialing(rabbitURL, queueName)
	taskService := service.New(publisher, log.Default())
	handler := httpapi.New(taskService)

	addr := ":" + port
	log.Printf("tasks service started on %s", addr)
	if err := http.ListenAndServe(addr, handler.Routes()); err != nil {
		log.Fatalf("tasks service error: %v", err)
	}
}

package main

import (
	"log"
	"net/http"

	"pz14-task-queue/internal/env"
	amqpqueue "pz14-task-queue/services/tasks/internal/amqp"
	httpapi "pz14-task-queue/services/tasks/internal/http"
	taskjobs "pz14-task-queue/services/tasks/internal/jobs"
)

func main() {
	rabbitURL := env.String("RABBIT_URL", "amqp://guest:guest@localhost:5672/")
	taskQueue := env.String("TASK_QUEUE", "task_jobs")
	dlqQueue := env.String("DLQ_QUEUE", "task_jobs_dlq")
	port := env.String("TASKS_PORT", "8082")

	conn, ch, err := amqpqueue.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("connect rabbitmq: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	if err := amqpqueue.DeclareQueues(ch, taskQueue, dlqQueue); err != nil {
		log.Fatalf("declare queues: %v", err)
	}

	publisher := amqpqueue.NewPublisher(ch)
	service := taskjobs.NewService(publisher, taskQueue, nil)
	handler := httpapi.NewHandler(service)

	addr := ":" + port
	log.Printf("tasks service listening on %s", addr)
	if err := http.ListenAndServe(addr, handler.Routes()); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

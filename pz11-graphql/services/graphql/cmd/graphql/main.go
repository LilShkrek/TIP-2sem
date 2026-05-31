package main

import (
	"log"
	"net/http"
	"os"

	"example.com/pz11-graphql/services/graphql/graph"
	"example.com/pz11-graphql/services/graphql/internal/repository"
	"example.com/pz11-graphql/services/graphql/internal/service"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	taskRepo := repository.NewMemoryTaskRepository()
	taskService := service.NewTaskService(taskRepo)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: graph.NewResolver(taskService),
	}))

	http.Handle("/", playground.Handler("Task GraphQL Playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("GraphQL Playground: http://localhost:%s/", port)
	log.Printf("GraphQL endpoint: http://localhost:%s/query", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

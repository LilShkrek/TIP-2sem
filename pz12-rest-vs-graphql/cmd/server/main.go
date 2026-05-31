package main

import (
	"log"
	"net/http"
	"os"

	"example.com/pz12-rest-vs-graphql/graph"
	"example.com/pz12-rest-vs-graphql/internal/rest"
	"example.com/pz12-rest-vs-graphql/internal/task"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8082"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	taskRepo := task.NewMemoryRepository(task.DefaultSeed())
	taskService := task.NewService(taskRepo)

	mux := http.NewServeMux()
	rest.NewHandler(taskService).Register(mux)

	graphqlServer := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: graph.NewResolver(taskService),
	}))
	mux.Handle("/", playground.Handler("Task GraphQL Playground", "/query"))
	mux.Handle("/query", graphqlServer)

	log.Printf("REST API: http://localhost:%s/v1/tasks", port)
	log.Printf("GraphQL Playground: http://localhost:%s/", port)
	log.Printf("GraphQL endpoint: http://localhost:%s/query", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

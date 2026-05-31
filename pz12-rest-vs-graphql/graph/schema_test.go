package graph

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/pz12-rest-vs-graphql/internal/task"
	"github.com/99designs/gqlgen/graphql/handler"
)

func TestGraphQLTaskScenario(t *testing.T) {
	service := task.NewService(task.NewMemoryRepository(task.DefaultSeed()))
	server := handler.NewDefaultServer(NewExecutableSchema(Config{
		Resolvers: NewResolver(service),
	}))

	list := graphqlRequest(t, server, `query { tasks { id title done } }`, nil)
	if len(list.Errors) != 0 {
		t.Fatalf("list errors: %+v", list.Errors)
	}
	if _, ok := list.Data["tasks"]; !ok {
		t.Fatal("tasks field is missing")
	}

	create := graphqlRequest(t, server, `
		mutation Create($input: CreateTaskInput!) {
			createTask(input: $input) { id title description done }
		}
	`, map[string]any{
		"input": map[string]any{
			"title":       "Новая задача",
			"description": "GraphQL",
		},
	})
	if len(create.Errors) != 0 {
		t.Fatalf("create errors: %+v", create.Errors)
	}

	update := graphqlRequest(t, server, `
		mutation Update($id: ID!, $input: UpdateTaskInput!) {
			updateTask(id: $id, input: $input) { id title description done }
		}
	`, map[string]any{
		"id": "t_003",
		"input": map[string]any{
			"done": true,
		},
	})
	if len(update.Errors) != 0 {
		t.Fatalf("update errors: %+v", update.Errors)
	}
}

func TestGraphQLNotFoundReturnsErrors(t *testing.T) {
	service := task.NewService(task.NewMemoryRepository(task.DefaultSeed()))
	server := handler.NewDefaultServer(NewExecutableSchema(Config{
		Resolvers: NewResolver(service),
	}))

	response := graphqlRequest(t, server, `query { task(id: "unknown") { id title } }`, nil)
	if len(response.Errors) == 0 {
		t.Fatal("expected GraphQL errors")
	}
}

type graphqlResponse struct {
	Data   map[string]any     `json:"data"`
	Errors []graphqlErrorBody `json:"errors"`
}

type graphqlErrorBody struct {
	Message string `json:"message"`
}

func graphqlRequest(t *testing.T, server http.Handler, query string, variables map[string]any) graphqlResponse {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	var response graphqlResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}

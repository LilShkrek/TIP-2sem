package student

import (
	"context"
	"testing"

	"example.com/pz2-grpc/gen/studentpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServicePing(t *testing.T) {
	service := NewService(NewRepository())

	resp, err := service.Ping(context.Background(), &studentpb.PingRequest{Message: "hello grpc"})
	if err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}

	want := "Server received: hello grpc"
	if resp.GetMessage() != want {
		t.Fatalf("Ping message = %q, want %q", resp.GetMessage(), want)
	}
}

func TestServiceGetStudentByID(t *testing.T) {
	service := NewService(NewRepository())

	resp, err := service.GetStudentByID(context.Background(), &studentpb.GetStudentRequest{Id: 1})
	if err != nil {
		t.Fatalf("GetStudentByID returned error: %v", err)
	}

	if resp.GetStudent().GetFullName() != "Иванов Иван Иванович" {
		t.Fatalf("unexpected student: %v", resp.GetStudent())
	}
}

func TestServiceGetStudentByIDNotFound(t *testing.T) {
	service := NewService(NewRepository())

	_, err := service.GetStudentByID(context.Background(), &studentpb.GetStudentRequest{Id: 999})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetStudentByID code = %v, want %v", status.Code(err), codes.NotFound)
	}
}

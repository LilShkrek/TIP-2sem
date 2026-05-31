package student

import (
	"context"
	"errors"

	"example.com/pz2-grpc/gen/studentpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	studentpb.UnimplementedStudentServiceServer
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Ping(ctx context.Context, req *studentpb.PingRequest) (*studentpb.PingResponse, error) {
	msg := req.GetMessage()
	if msg == "" {
		msg = "ping"
	}

	return &studentpb.PingResponse{
		Message: "Server received: " + msg,
	}, nil
}

func (s *Service) GetStudentByID(ctx context.Context, req *studentpb.GetStudentRequest) (*studentpb.GetStudentResponse, error) {
	id := req.GetId()
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid student id")
	}

	st, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrStudentNotFound) {
			return nil, status.Error(codes.NotFound, "student not found")
		}
		return nil, status.Error(codes.Internal, "failed to get student")
	}

	return &studentpb.GetStudentResponse{
		Student: st,
	}, nil
}

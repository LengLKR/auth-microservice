package transport

import (
	"context"

    pb "github.com/LengLKR/auth-microservice/internal/transport/proto"
    "github.com/LengLKR/auth-microservice/internal/service"
    "google.golang.org/grpc"
)

// Server implements pb.AuthServiceServer
type Server struct {
    authSvc *service.AuthService
    pb.UnimplementedAuthServiceServer
}

// NewServer คืน instance ของ Server
func NewServer(svc *service.AuthService) pb.AuthServiceServer {
    return &Server{authSvc: svc}
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
    token, err := s.authSvc.Register(ctx, req.Email, req.Password)
    if err != nil {
        return nil, err
    }
    return &pb.AuthResponse{Token: token}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
    token, err := s.authSvc.Login(ctx, req.Email, req.Password)
    if err != nil {
        return nil, err
    }
    return &pb.AuthResponse{Token: token}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.Empty, error) {
    if err := s.authSvc.Logout(ctx, req.Token); err != nil {
		return nil, err
	}
    return &pb.Empty{}, nil
}

// RegisterAuthServiceServer ช่วย register ใน main.go
func RegisterAuthServiceServer(grpcServer *grpc.Server, srv pb.AuthServiceServer) {
    pb.RegisterAuthServiceServer(grpcServer, srv)
}

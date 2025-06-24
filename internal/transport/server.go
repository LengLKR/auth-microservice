package transport

import (
	"context"
    "time"

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

//ListUsers
func (s *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error){
    users, total, err := s.authSvc.ListUsers(ctx, req.FilterName, req.FilterEmail, int(req.Page), int(req.Size))
    if err != nil {
        return nil, err
    }
    pbUsers := make([]*pb.User, len(users))
    for i, u := range users {
        pbUsers[i] = &pb.User{
            Id:         u.ID,
            Email:      u.Email,
            CreatedAt:  u.CreatedAt.Format(time.RFC3339),
        }
    }
    return &pb.ListUsersResponse{Users: pbUsers, TotalCount: int32(total)}, nil
}

//GetProfie
func (s *Server) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.User, error) {
    u, err := s.authSvc.GetProfile(ctx, req.Id)
    if err != nil {
        return nil, err
    }
    return &pb.User{
        Id:         u.ID,
        Email:      u.Email,
        CreatedAt:  u.CreatedAt.Format(time.RFC3339),
    }, nil
}

//UpdaeProfile
func (s *Server) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.User, error){
    u, err := s.authSvc.UpdateProfile(ctx, req.Id, req.Email)
    if err != nil {
        return nil, err
    }
    return &pb.User{
        Id:         u.ID,
        Email:      u.Email,
        CreatedAt:  u.CreatedAt.Format(time.RFC3339),
    }, nil
}

//DeleteProfile
func (s *Server) DeleteProfile(ctx context.Context, req *pb.DeleteProfileRequest) (*pb.Empty,error){
    if err := s.authSvc.DeleteProfile(ctx, req.Id); err != nil {
        return nil, err
    }
    return &pb.Empty{}, nil
}

// RequestPasswordReset สั่งสร้าง reset token
func (s *Server) RequestPasswordReset(ctx context.Context, req *pb.PasswordResetRequest) (*pb.Empty, error) {
    if err := s.authSvc.RequestPasswordReset(ctx, req.Email); err != nil {
        return nil, err
    }
    return &pb.Empty{}, nil
}

// ResetPassword ตรวจ token และรีเซ็ตรหัสผ่าน
func (s *Server) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.Empty, error) {
    if err := s.authSvc.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
        return nil, err
    }
    return &pb.Empty{}, nil
}

// RegisterAuthServiceServer ช่วย register ใน main.go
func RegisterAuthServiceServer(grpcServer *grpc.Server, srv pb.AuthServiceServer) {
    pb.RegisterAuthServiceServer(grpcServer, srv)
}



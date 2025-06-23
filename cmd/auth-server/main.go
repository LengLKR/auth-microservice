package main

import (
	"log"
	"net"
	"github.com/joho/godotenv"
    "github.com/LengLKR/auth-microservice/config"
    "github.com/LengLKR/auth-microservice/internal/repository"
    "github.com/LengLKR/auth-microservice/internal/service"
    "github.com/LengLKR/auth-microservice/internal/transport"
    "google.golang.org/grpc"
)

func main(){
	// โหลด .env (ถ้าใช้ godotenv)
	_= godotenv.Load()

	// โหลด config จาก env
	cfg := config.Load()

	// เชื่อมต่อ MongoDB
	client, err := config.InitMongo(cfg)
	if err != nil {
	log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	db := client.Database(cfg.MongoDatabase)

	// สร้าง repository & service
	userCol := db.Collection("users")
	userRepo := repository.NewMongoUserRepository(userCol)

	//สร้าง token repository สำหรับ Logout blacklist
	tokenCol := db.Collection("invalidated_tokens")
	tokenRepo := repository.NewMongoTokenRepository(tokenCol)
	
	// สร้าง AuthService พร้อมทั้ง userRepo, tokenRepo, และ jwtSecret
	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret)


	// สร้าง gRPC server และ register
	grpcServer := grpc.NewServer()
	transport.RegisterAuthServiceServer(grpcServer, transport.NewServer(authSvc))

	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("gRPC server listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
package config

import (
	"context"
	"fmt"
	"os"
	"time"
	
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
) 

// config เก็บค่าต่างๆ  จาก env
type Config struct {

	MongoURI string
	MongoDatabase string
	JWTSecret string

}

//Load อ่านค่าจาก enviroment varibles
func Load() *Config {

	return &Config{
		MongoURI:      os.Getenv("MONGO_URI"),      // เช่น mongodb://localhost:27017
		MongoDatabase: os.Getenv("MONGO_DATABASE"), // เช่น authdb
		JWTSecret:     os.Getenv("JWT_SECRET"),     // secret สำหรับเซ็น JWT
	}
}

// InitMongo เชื่อมต่อ MongoDB client
func InitMongo(cfg *Config) (*mongo.Client, error){
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOpts)

	if err != nil {
		return nil, fmt.Errorf("mongo connect error: %w", err)
	}
	return client, nil
}
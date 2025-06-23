package repository

import (

	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" 

)

// TokenRepository interface สำหรับจัดการ blacklist tokens
type TokenRepository interface {

	Blacklist(token string, expiresAt time.Time) error
	IsBlacklisted(token string) (bool, error)

}

type mongoTokenRepo struct {

	col *mongo.Collection

}

// NewMongoTokenRepository สร้าง instance และตั้ง TTL index
func NewMongoTokenRepository(col *mongo.Collection) TokenRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"expiresAt": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	return &mongoTokenRepo{col: col}
}

func (r *mongoTokenRepo) Blacklist(token string, expiresAt time.Time) error {
	_, err := r.col.InsertOne(context.Background(), bson.M{
	"token":     token,
	"expiresAt": expiresAt,
	})
	return err
}

func (r *mongoTokenRepo) IsBlacklisted(token string) (bool, error) {
	count, err := r.col.CountDocuments(context.Background(), bson.M{"token": token})
	return count > 0, err
}
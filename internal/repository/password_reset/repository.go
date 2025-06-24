package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//PasswordResetRepository  จัดการสร้าง-ตรวจสอบ-ลบ reset token
type PasswordResetRepository interface {
	Create(token, userID string, expiresAt time.Time) error
	Verify(token string) (string, error)
	Delete(token string) error
}

type mongoResetRepo struct {
	col *mongo.Collection
}

func NewMongoResetRepo(col *mongo.Collection) PasswordResetRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//index บน token และ TTL บน expiresAT
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:	bson.M{"token": 1},
		Options: options.Index().SetUnique(true),
	})
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"expiresAT": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	return &mongoResetRepo{col: col}
}

func (r *mongoResetRepo) Create(token, userID string, expiresAt time.Time) error{
	_, err := r.col.InsertOne(context.Background(), bson.M{
		"token":	 token,
		"userID":	 userID,
		"expiresAT": expiresAt,
	})
	return err
}

func (r *mongoResetRepo) Verify(token string) (string, error){
	var doc struct{ UserID string}
	err := r.col.FindOne(context.Background(), bson.M{"token": token}).Decode(&doc)
	return doc.UserID, err
}

func (r *mongoResetRepo) Delete(token string) error {
	_, err := r.col.DeleteOne(context.Background(), bson.M{"token": token})
	return err
}
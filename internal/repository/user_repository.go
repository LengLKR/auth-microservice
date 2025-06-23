package repository

import (
	
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" 
	"github.com/LengLKR/auth-microservice/internal/domain"

)

//Userrepository is inter for user data access
type UserRepository interface {

	Create(u *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindAll(filterName, filterEmail string, page, size int ) ([]*domain.User, int64, error)
	FindByID(id string) (*domain.User, error)
	Update( u *domain.User) error
	SoftDelete(id string) error
}

//mongoUserRepo is MongoDB implementtation of Userrepository
type mongoUserRepo  struct {
	col *mongo.Collection
}

// NewMongoUserRepository constructs a MongoDB-backed repository
func NewMongoUserRepository(col *mongo.Collection) UserRepository {
    // สร้าง unique index บน email
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys:    bson.M{"email": 1},
        Options: options.Index().SetUnique(true),
    })
    if err != nil {
        // ตรงนี้เลือกจะ panic หรือ log ก็ได้ ขึ้นกับแนวทางทีม
        panic("failed to create index on users.email: " + err.Error())
    }

    return &mongoUserRepo{col: col}
}

func (r *mongoUserRepo) Create(u *domain.User) error {
	u.CreatedAt = time.Now()
	_, err := r.col.InsertOne(context.Background(), u)
	return err
}

func (r *mongoUserRepo) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	err := r.col.FindOne(context.Background(), bson.M{"email": email}).Decode(&u)
	if err == mongo.ErrNoDocuments {
	return nil, errors.New("user not found")
	}
	return &u, err
}

// FindAll return filtered users and total count.
func (r *mongoUserRepo) FindAll(filterName, filterEmail string, page, size int) ([]*domain.User, int64, error) {

	ctx := context.Background()
	filter := bson.M{"deletedAt": bson.M{"$exists": false}}
	if filterName != "" {
		filter["name"] = bson.M{"$regex": filterName, "$options": "i"}
	}
	if filterEmail != "" {
		filter["email"] = bson.M{"$regex": filterEmail, "$options": "i"}
	}

	total, err := r.col.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }

    opts := options.Find().
        SetSkip(int64((page - 1) * size)).
        SetLimit(int64(size))
	
	cursor, err := r.col.Find(ctx, filter, opts)

	if err != nil {
		return nil, 0 , err
	}

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0 , err
	}
	return  users, total, nil
}

// FindByID returns a single user by ID excluding soft-deleted.
func (r *mongoUserRepo) FindByID(id string) (*domain.User, error) {
    ctx := context.Background()
    filter := bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}
    var u domain.User
    err := r.col.FindOne(ctx, filter).Decode(&u)
    if err == mongo.ErrNoDocuments {
        return nil, errors.New("user not found")
    }
    return &u, err
}

// Update modifies allowed fields of a user.
func (r *mongoUserRepo) Update(u *domain.User) error {
    _, err := r.col.UpdateOne(
        context.Background(),
        bson.M{"_id": u.ID},
        bson.M{"$set": bson.M{
            "email": u.Email,
            // เพิ่มฟิลด์อื่นๆ ที่อนุญาตตามต้องการ
        }},
    )
    return err
}

// SoftDelete marks a user as deleted by setting deletedAt.
func (r *mongoUserRepo) SoftDelete(id string) error {
    _, err := r.col.UpdateOne(
        context.Background(),
        bson.M{"_id": id},
        bson.M{"$set": bson.M{"deletedAt": time.Now()}},
    )
    return err
}
// internal/domain/user.go
package domain

import "time"

// User  model
type User struct {
	ID           string    `bson:"_id,omitempty"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"password_hash"`
	CreatedAt    time.Time `bson:"created_at"` 
	Name         string     `bson:"name,omitempty"`       
	DeletedAt    *time.Time `bson:"deletedAt,omitempty"`  // สำหรับ soft delete
}
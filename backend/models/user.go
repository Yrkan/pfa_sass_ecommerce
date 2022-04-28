package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Username string               `json:"username,omitempty" bson:"username,omitempty" validate:"required"`
	Password string               `json:"password,omitempty" bson:"password,omitempty" validate:"required"`
	Email    string               `json:"email,omitempty" bson:"email,omitempty" validate:"required"`
	FullName string               `json:"full_name,omitempty" bson:"full_name,omitempty" validate:"required"`
	Stores   []primitive.ObjectID `json:"stores,omitempty" bson:"stores,omitempty" validate:"required"`
}

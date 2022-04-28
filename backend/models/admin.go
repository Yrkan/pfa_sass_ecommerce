package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Admin struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username,omitempty" bson:"username,omitempty" validate:"required"`
	Password string             `json:"password,omitempty" bson:"password,omitempty" validate:"required"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty" validate:"required"`
}

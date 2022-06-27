package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Store struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty" validate:"required"`
	Owner primitive.ObjectID `json:"owner,omitempty" bson:"owner,omitempty"`
}

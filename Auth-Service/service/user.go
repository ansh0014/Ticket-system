package service

import (
	"context"
	"time"

	"github.com/ansh0014/auth/config"
	"github.com/ansh0014/auth/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(email, passwordHash string) error {
	user := model.User{
		ID:        primitive.NewObjectID().Hex(),
		Email:     email,
		Password:  passwordHash,
		IsActive:  false,
		CreatedAt: time.Now(),
	}
	_, err := config.MongoDB.Collection("users").InsertOne(context.Background(), user)
	return err
}

func ActivateUser(email string) error {
	_, err := config.MongoDB.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"email": email},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	return err
}

package controllers

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/yrkan/pfa_sass_ecommerce/backend/config"
	"github.com/yrkan/pfa_sass_ecommerce/backend/models"
	"github.com/yrkan/pfa_sass_ecommerce/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gofiber/fiber/v2"
)

func GetAllUsers(c *fiber.Ctx) error {

	// Check authorization
	authorized := false
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	tokenAdminId := claims["admin_id"]

	if tokenAdminId != nil {
		authorized = true
	}

	if !authorized {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Authorized
	usersCollection := config.MI.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var users []models.User

	filter := bson.M{}
	findOptions := options.Find()

	// Search
	if s := c.Query("s"); s != "" {
		filter = bson.M{
			"$or": []bson.M{
				{
					"username": bson.M{
						"$regex": primitive.Regex{
							Pattern: s,
							Options: "i",
						},
					},
				},
			},
		}
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limitVal, _ := strconv.Atoi(c.Query("limit", "10"))
	var limit int64 = int64(limitVal)

	total, _ := usersCollection.CountDocuments(ctx, filter)

	findOptions.SetSkip((int64(page) - 1) * limit)
	findOptions.SetLimit(limit)

	// Find users
	cursor, err := usersCollection.Find(ctx, filter, findOptions)

	// Error Handeling
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Users not found",
			"error":   err,
		})
	}

	defer cursor.Close(ctx)

	// Success
	for cursor.Next(ctx) {
		var user models.User
		cursor.Decode(&user)
		users = append(users, user)
	}

	last := float64(total / limit)
	if last < 1 && total > 0 {
		last = 1
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":      users,
		"total":     total,
		"page":      page,
		"last_page": last,
		"limit":     limit,
	})

}

func GetSingleUser(c *fiber.Ctx) error {
	// Check authorization
	authorized := false
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	userId, err := primitive.ObjectIDFromHex(c.Params("userId"))
	tokenUserId := claims["user_id"]
	tokenAdminId := claims["admin_id"]

	if tokenAdminId != nil {
		authorized = true
	} else if tokenUserId != nil {
		if tokenUserId == c.Params("userId") {
			authorized = true
		}
	}

	if !authorized {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Bad request
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Bad request",
			"error":   err,
		})
	}

	usersCollection := config.MI.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User

	// Not Found
	findResult := usersCollection.FindOne(ctx, bson.M{"_id": userId})
	if err := findResult.Err(); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
			"error":   err,
		})
	}

	err = findResult.Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
			"error":   err,
		})
	}

	// Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    user,
		"success": true,
	})
}

func CreateUser(c *fiber.Ctx) error {
	// Check authorization
	authorized := false
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	tokenAdminId := claims["admin_id"]

	if tokenAdminId != nil {
		authorized = true
	}

	if !authorized {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	usersCollection := config.MI.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	user := new(models.User)

	// Bad request
	if err := c.BodyParser(user); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"sucess":  false,
			"message": "Failed to parse the request body",
			"error":   err,
		})
	}

	// Validation
	validate := utils.NewValidator()
	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": utils.ValidationErrors(err),
		})
	}

	// Hash password
	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create user",
			"error":   err,
		})
	}
	user.Password = hashed

	// Attempt insert
	result, err := usersCollection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create user",
			"error":   err,
		})
	}

	// Success
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
		"message": "User created successfully",
	})

}

func UpdateUser(c *fiber.Ctx) error {
	// Check authorization
	authorized := false
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	tokenUserId := claims["user_id"]
	tokenAdminId := claims["admin_id"]

	if tokenAdminId != nil {
		authorized = true
	} else if tokenUserId != nil {
		if tokenUserId == c.Params("userId") {
			authorized = true
		}
	}

	if !authorized {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	usersCollection := config.MI.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user := new(models.User)

	// Bad request
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Bad request",
			"error":   err,
		})
	}

	userId, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
			"error":   err,
		})
	}

	// Update document
	update := bson.M{
		"$set": user,
	}

	_, err = usersCollection.UpdateOne(ctx, bson.M{"_id": userId}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update user",
			"error":   err.Error(),
		})
	}

	// Success
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User updated successfully",
	})

}

func DeleteUser(c *fiber.Ctx) error {
	// Check authorization
	authorized := false
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	tokenUserId := claims["user_id"]
	tokenAdminId := claims["admin_id"]

	if tokenAdminId != nil {
		authorized = true
	} else if tokenUserId != nil {
		if tokenUserId == c.Params("userId") {
			authorized = true
		}
	}

	if !authorized {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Authorized
	usersCollection := config.MI.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userId, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
			"error":   err,
		})
	}

	_, err = usersCollection.DeleteOne(ctx, bson.M{"_id": userId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete user",
			"error":   err,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})

}

package controllers

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/yrkan/pfa_sass_ecommerce/backend/config"
	"github.com/yrkan/pfa_sass_ecommerce/backend/models"
	"github.com/yrkan/pfa_sass_ecommerce/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func getUserByUsername(username string) (*models.User, error) {
	usersCollection := config.MI.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User

	findResult := usersCollection.FindOne(ctx, bson.M{"username": username})
	err := findResult.Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, err

}

func Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input LoginInput
	var user *models.User

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request",
		})
	}

	username := input.Username
	pass := input.Password

	// Check user exists
	user, err := getUserByUsername(username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid username or password",
		})
	}

	// Validate password correct
	if !utils.CheckPasswordHash(pass, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid username or password",
		})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Success login",
		"data":    signedToken,
	})
}

func Register(c *fiber.Ctx) error {
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

	// Check user exists
	exists, _ := getUserByUsername(user.Username)
	if exists != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Username already in use",
		})
	}

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

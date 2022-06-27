package controllers

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/yrkan/pfa_sass_ecommerce/backend/config"
	"github.com/yrkan/pfa_sass_ecommerce/backend/models"
	"github.com/yrkan/pfa_sass_ecommerce/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAllStores(c *fiber.Ctx) error {
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
	storesCollection := config.MI.DB.Collection("stores")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stores []models.Store

	filter := bson.M{}
	findOptions := options.Find()

	// Search
	if s := c.Query("s"); s != "" {
		filter = bson.M{
			"$or": []bson.M{
				{
					"name": bson.M{
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

	total, _ := storesCollection.CountDocuments(ctx, filter)

	findOptions.SetSkip((int64(page) - 1) * limit)
	findOptions.SetLimit(limit)

	// Find stores
	cursor, err := storesCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Stores not found",
			"error":   err,
		})
	}
	defer cursor.Close(ctx)

	// Success
	for cursor.Next(ctx) {
		var store models.Store
		cursor.Decode(&store)
		stores = append(stores, store)
	}

	last := float64(total / limit)
	if last < 1 && total > 0 {
		last = 1
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":      stores,
		"total":     total,
		"page":      page,
		"last_page": last,
		"limit":     limit,
	})

}

func GetSingleStore(c *fiber.Ctx) error {
	// Bad request
	storeId, err := primitive.ObjectIDFromHex(c.Params("storeId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Bad request",
			"error":   err,
		})
	}

	storesCollection := config.MI.DB.Collection("stores")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var store models.Store

	// Not Found
	findOptions := options.FindOne().SetProjection(bson.D{{"owner", 0}})
	findResult := storesCollection.FindOne(ctx, bson.M{"_id": storeId}, findOptions)
	if err := findResult.Err(); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Store not found",
			"error":   err,
		})
	}

	err = findResult.Decode(&store)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Store not found",
			"error":   err,
		})
	}

	// Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    store,
		"success": true,
	})
}

func CreateStore(c *fiber.Ctx) error {
	// Check authorization
	authorized := false
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	tokenAdminId := claims["admin_id"]
	tokenUserId := claims["user_id"]

	if tokenAdminId != nil || tokenUserId != nil {
		authorized = true
	}

	if !authorized {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Init store
	storesCollection := config.MI.DB.Collection("stores")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store := new(models.Store)
	if primitive.IsValidObjectID(tokenUserId.(string)) {
		store.Owner, _ = primitive.ObjectIDFromHex(tokenUserId.(string))
	} else if primitive.IsValidObjectID(tokenAdminId.(string)) {
		store.Owner, _ = primitive.ObjectIDFromHex(tokenAdminId.(string))
	}

	// Bad request
	if err := c.BodyParser(store); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"sucess":  false,
			"message": "Failed to parse the request body",
			"error":   err,
		})
	}

	// Validation
	validate := utils.NewValidator()
	if err := validate.Struct(store); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": utils.ValidationErrors(err),
		})
	}

	// Attempt insert
	result, err := storesCollection.InsertOne(ctx, store)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create the store",
			"error":   err,
		})
	}

	// Update user's stores list
	if tokenUserId != nil {
		usersCollection := config.MI.DB.Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		update := bson.M{
			"$push": bson.M{"stores": result.InsertedID},
		}
		userId, _ := primitive.ObjectIDFromHex(tokenUserId.(string))
		_, err := usersCollection.UpdateOne(ctx, bson.M{"_id": userId}, update, options.Update().SetUpsert(true))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update user",
				"error":   err.Error(),
			})
		}
	}

	// Success
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
		"message": "Store created successfully",
	})
}

func DeleteStore(c *fiber.Ctx) error {

	storeId, err := primitive.ObjectIDFromHex(c.Params("storeId"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Store not found",
			"error":   err,
		})
	}

	// Check authorization
	authorized := false
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	tokenUserId := claims["user_id"]
	tokenAdminId := claims["admin_id"]

	if tokenAdminId != nil {
		authorized = true
	} else if tokenUserId != nil {
		usersCollection := config.MI.DB.Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		var user models.User

		userId, _ := primitive.ObjectIDFromHex(tokenUserId.(string))
		findResult := usersCollection.FindOne(ctx, bson.M{"_id": userId})
		if err := findResult.Err(); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
				"error":   err,
			})
		}
		// Check user own store
		if err := findResult.Decode(&user); err == nil {
			for _, v := range user.Stores {
				if v == storeId {
					authorized = true
					// Remove store ref from stores array
					updates := bson.M{
						"$pull": bson.M{"stores": storeId},
					}
					usersCollection.UpdateByID(ctx, userId, updates)
				}
			}
		}

	}
	if !authorized {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Authorized
	storesCollection := config.MI.DB.Collection("stores")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err = storesCollection.DeleteOne(ctx, bson.M{"_id": storeId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete store",
			"error":   err,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Store deleted successfully",
	})
}

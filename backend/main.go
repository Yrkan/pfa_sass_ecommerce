package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/yrkan/pfa_sass_ecommerce/backend/config"
	"github.com/yrkan/pfa_sass_ecommerce/backend/models"
	"github.com/yrkan/pfa_sass_ecommerce/backend/routes"
	"github.com/yrkan/pfa_sass_ecommerce/backend/utils"
)

// If there's no admin (first setup) ask for an admin
func setupAdminIfNotExist() {
	// Check if there is already an admin
	adminsCollection := config.MI.DB.Collection("admins")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	countAdmins, _ := adminsCollection.CountDocuments(ctx, fiber.Map{})

	// No admins ask for one to be created
	if countAdmins == 0 {
		var answer string
		fmt.Println("WARRNING: There is no admin, configure a new admin ?")
		fmt.Scan(&answer)

		if strings.ToUpper(answer) == "Y" {
			admin := new(models.Admin)
			fmt.Println("Enter the admin's username:")
			fmt.Scan(&answer)
			admin.Username = answer

			fmt.Println("Enter the admin's password:")
			fmt.Scan(&answer)
			hashed, err := utils.HashPassword(answer)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error hashing password")
				os.Exit(1)
			}
			admin.Password = hashed

			fmt.Println("Enter the admin's email:")
			fmt.Scan(&answer)
			admin.Email = answer

			result, err := adminsCollection.InsertOne(ctx, admin)
			if err != nil {
				fmt.Println("Error inserting the new admin to the database")
				fmt.Println("Details: ", err)
				os.Exit(2)
			}

			// Success
			fmt.Println("Admin created with success")
			fmt.Println(result.InsertedID)

		} else {
			fmt.Println("there's nothing to do")
			os.Exit(0)
		}
	}

}

// Setup the routes by grouping them
func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Root Page",
		})
	})

	api := app.Group("/api")

	routes.UsersRoute(api.Group("/users"))
	routes.AuthRoutes(api.Group("/auth"))
	routes.StoresRoutes(api.Group("/stores"))
}

func main() {
	config.ConnectDB()

	setupAdminIfNotExist()

	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	app := fiber.New(fiber.Config{
		Prefork: false,
	})

	app.Use(cors.New())
	app.Use(logger.New())

	setupRoutes(app)

	port := os.Getenv("PORT")
	err := app.Listen(":" + port)
	if err != nil {
		log.Fatal("Error app failed to start")
		panic(err)
	}
}

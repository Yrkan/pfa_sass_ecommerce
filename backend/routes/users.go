package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yrkan/pfa_sass_ecommerce/backend/controllers"
	middlewares "github.com/yrkan/pfa_sass_ecommerce/backend/middleware"
)

func UsersRoute(route fiber.Router) {
	// Get all users
	route.Get("/", middlewares.Protected(), controllers.GetAllUsers)
	// Get a single user
	route.Get("/:userId", middlewares.Protected(), controllers.GetSingleUser)
	// Create a user as admin
	route.Post("/", middlewares.Protected(), controllers.CreateUser)
	// Update a user as admin
	route.Patch("/:userId", middlewares.Protected(), controllers.UpdateUser)
	// Delete user
	route.Delete(":/userId", middlewares.Protected(), controllers.DeleteUser)
}

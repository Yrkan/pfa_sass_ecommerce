package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yrkan/pfa_sass_ecommerce/backend/controllers"
	middlewares "github.com/yrkan/pfa_sass_ecommerce/backend/middleware"
)

func StoresRoutes(route fiber.Router) {
	// Create a store
	route.Post("/", middlewares.Protected(), controllers.CreateStore)
	// Get all stores
	route.Get("/", middlewares.Protected(), controllers.GetAllStores)
	// Get single store
	route.Get("/:storeId", controllers.GetSingleStore)
	// Delete single store
	route.Delete("/:storeId", middlewares.Protected(), controllers.DeleteStore)
}

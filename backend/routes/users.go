package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yrkan/pfa_sass_ecommerce/backend/controllers"
)

func UsersRoute(route fiber.Router) {
	route.Get("/", controllers.GetAllUsers)
	route.Get("/:userId", controllers.GetSingleUser)
	route.Post("/", controllers.CreateUser)
}

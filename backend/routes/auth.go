package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yrkan/pfa_sass_ecommerce/backend/controllers"
)

func AuthRoutes(route fiber.Router) {
	route.Post("/login", controllers.Login)
	route.Post("/register", controllers.Register)
}

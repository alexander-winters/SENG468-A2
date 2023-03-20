package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/alexander-winters/SENG468-A2/mongo"
)

func main() {
	// Initialize a new Fiber app
	app := fiber.New()

	// Connect to MongoDB
	client := mongo.GetMongoClient()

	// Set up a simple route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Set up the routes for users
	app.Post("/api/users", CreateUser)
	app.Get("/api/users/:id", GetUser)
	app.Put("/api/users/:id", UpdateUser)
	app.Delete("/api/users/:id", DeleteUser)
	app.Get("/api/users", ListUsers)

	// Set up the routes for posts
	app.Post("/api/posts", CreatePost)
	app.Get("/api/posts/:id", GetPost)
	app.Put("/api/posts/:id", UpdatePost)
	app.Delete("/api/posts/:id", DeletePost)
	app.Get("/api/posts", ListPosts)

	// Set up the routes for comments
	app.Post("/api/comments", CreateComment)
	app.Get("/api/comments/:id", GetComment)
	app.Put("/api/comments/:id", UpdateComment)
	app.Delete("/api/comments/:id", DeleteComment)
	app.Get("/api/comments", ListComments)

	// Set up the routes for notifications
	app.Post("/api/notifications", CreateNotification)
	app.Get("/api/notifications/:id", GetNotification)
	app.Put("/api/notifications/:id", UpdateNotification)
	app.Delete("/api/notifications/:id", DeleteNotification)
	app.Get("/api/notifications", ListNotifications)

	// Start the server on the specified port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}

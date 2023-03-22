package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/alexander-winters/SENG468-A2/server/routes"
)

// Connect to MongoDB
// var client = mymongo.GetMongoClient()

func main() {
	// Initialize a new Fiber app
	app := fiber.New()

	// Set up a simple route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Set up the routes for users
	app.Post("/api/users", routes.CreateUser)
	app.Get("/api/users/:id", routes.GetUser)
	app.Put("/api/users/:id", routes.UpdateUser)
	app.Delete("/api/users/:id", routes.DeleteUser)
	app.Get("/api/users", routes.ListUsers)

	// Set up the routes for posts
	app.Post("/api/posts", routes.CreatePost)
	app.Get("/api/posts/:id", routes.GetPost)
	app.Put("/api/posts/:id", routes.UpdatePost)
	app.Delete("/api/posts/:id", routes.DeletePost)
	app.Get("/api/posts", routes.ListPosts)

	// Set up the routes for comments
	app.Post("/api/comments", routes.CreateComment)
	app.Get("/api/comments/:id", routes.GetComment)
	app.Put("/api/comments/:id", routes.UpdateComment)
	app.Delete("/api/comments/:id", routes.DeleteComment)
	app.Get("/api/comments", routes.ListComments)

	// Set up the routes for notifications
	app.Post("/api/notifications", routes.CreateNotification)
	app.Put("/api/notifications/:id", routes.MarkNotificationAsRead)
	app.Get("/api/notifications", routes.ListNotifications)

	// Set up the routes for reports
	app.Get("/api/reports/posts", routes.PostReports)
	app.Get("/api/reports/comments", routes.UserCommentReports)
	app.Get("/api/reports/likes", routes.LikeReports)

	// Start the server on the specified port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}

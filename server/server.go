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
	app.Get("/api/users/:username", routes.GetUser)
	app.Put("/api/users/:username", routes.UpdateUser)
	app.Delete("/api/users/:username", routes.DeleteUser)
	app.Get("/api/users", routes.ListUsers)

	// Set up the routes for posts
	app.Post("/api/users/:username/posts", routes.CreatePost)
	app.Get("/api/users/:username/posts/:post_number", routes.GetPost)
	app.Put("/api/users/:username/posts/:post_number", routes.UpdatePost)
	app.Delete("/api/users/:username/posts/:post_number", routes.DeletePost)
	app.Get("/api/users/:username/posts", routes.ListUserPosts)
	app.Get("/api/posts", routes.ListAllPosts)

	// Set up the routes for comments
	app.Post("/api/posts/:post_number/comments", routes.CreateComment)
	app.Get("/api/posts/:post_number/comments/:id", routes.GetComment)
	app.Put("/api/comments/:id", routes.UpdateComment)
	app.Delete("/api/comments/:id", routes.DeleteComment)
	app.Get("/api/posts/:post_number/comments", routes.ListComments)

	// Set up the routes for notifications
	app.Post("/api/notifications", routes.CreateNotification)
	app.Put("/api/notifications/:id", routes.MarkNotificationAsRead)
	app.Get("/api/notifications/:username", routes.ListNotifications)

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

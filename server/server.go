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
	app.Post("/user", routes.CreateUser)
	app.Get("/user/:username", routes.GetUser)
	app.Put("/user/:username", routes.UpdateUser)
	app.Delete("/user/:username", routes.DeleteUser)
	app.Get("/users", routes.ListUsers)

	// Set up the routes for posts
	app.Post("/user/:username/post", routes.CreatePost)
	app.Get("/user/:username/post/:post_number", routes.GetPost)
	app.Put("/user/:username/post/:post_number", routes.UpdatePost)
	app.Delete("/user/:username/post/:post_number", routes.DeletePost)
	app.Get("/users/:username/posts", routes.ListUserPosts)
	app.Get("/posts", routes.ListAllPosts)
	app.Put("/user/:username/post/:post_number/like", routes.LikePost)

	// Set up the routes for comments
	app.Post("user/:username/post/:post_number/comment", routes.CreateComment)
	app.Get("/user/:username/post/:post_number/comment", routes.GetComment)
	app.Put("user/:username/post/:post_number/comment", routes.UpdateComment)
	app.Delete("user/:username/post/:post_number/comment", routes.DeleteComment)
	app.Get("/post/:post_number/comments", routes.ListComments)
	app.Put("/user/:username/post/:post_number/comment/like", routes.LikeComment)

	// Set up the routes for reports
	app.Get("/reports/user/posts", routes.PostReport)
	app.Get("/reports/user/comments", routes.UserCommentReport)
	app.Get("/reports/user/likes", routes.LikeReport)

	// Start the server on the specified port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}

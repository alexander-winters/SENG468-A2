package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/alexander-winters/SENG468-A2/kafka-docker/kafka"
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
	app.Post("/users", routes.CreateUser)
	app.Get("/users/:username", routes.GetUser)
	app.Put("/users/:username", routes.UpdateUser)
	app.Delete("/users/:username", routes.DeleteUser)
	app.Get("/users", routes.ListUsers)

	// Set up the routes for posts
	app.Post("/users/:username/posts", routes.CreatePost)
	app.Get("/users/:username/posts/:post_number", routes.GetPost)
	app.Put("/users/:username/posts/:post_number", routes.UpdatePost)
	app.Delete("/users/:username/posts/:post_number", routes.DeletePost)
	app.Get("/users/:username/posts", routes.ListUserPosts)
	app.Get("/posts", routes.ListAllPosts)

	// Set up the routes for comments
	app.Post("/posts/:post_number/comments", routes.CreateComment)
	app.Get("/posts/:post_number/comments/:id", routes.GetComment)
	app.Put("/comments/:id", routes.UpdateComment)
	app.Delete("/comments/:id", routes.DeleteComment)
	app.Get("/posts/:post_number/comments", routes.ListComments)

	// Set up the routes for notifications
	app.Post("/notifications", routes.CreateNotification)
	app.Put("/notifications/:id", routes.MarkNotificationAsRead)
	app.Get("/notifications/:username", routes.ListNotifications)

	// Set up the routes for reports
	app.Get("/reports/user/posts", routes.PostReport)
	app.Get("/reports/user/comments", routes.UserCommentReport)
	app.Get("/reports/user/likes", routes.LikeReport)

	// Initialize Kafka producer and consumer
	kafkaBrokerURL := "kafka:9092"
	kafkaProducer := kafka.CreateKafkaProducer(kafkaBrokerURL)
	kafkaConsumer := kafka.CreateKafkaConsumer(kafkaBrokerURL)

	// Start the server on the specified port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}

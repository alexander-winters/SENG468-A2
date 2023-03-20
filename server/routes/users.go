package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/alexander-winters/SENG468-A2/mongo"
	"github.com/alexander-winters/SENG468-A2/mongo/models"
)

// CreateUser inserts a new user into the database
func CreateUser(c *fiber.Ctx) error {
	// Parse the request body into a struct
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the created time
	user.CreatedAt = time.Now()

	// Insert the user into the database
	res, err := mongo.Users.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert user into database",
		})
	}

	// Set the ID of the user and return it
	user.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(user)
}

// GetUser retrieves a user from the database by ID
func GetUser(c *fiber.Ctx) error {
	// Your implementation here
}

// UpdateUser updates a user in the database by ID
func UpdateUser(c *fiber.Ctx) error {
	// Your implementation here
}

// DeleteUser deletes a user from the database by ID
func DeleteUser(c *fiber.Ctx) error {
	// Your implementation here
}

// ListUsers retrieves all users from the database
func ListUsers(c *fiber.Ctx) error {
	// Your implementation here
}

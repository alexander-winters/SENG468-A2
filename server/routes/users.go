package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
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
	// Get the user ID from the request parameters
	userID := c.Params("id")

	// Convert the user ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Find the user in the database by ID
	var user models.User
	err = mongo.Users.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve user from database",
		})
	}

	return c.JSON(user)
}

// UpdateUser updates a user in the database by Username
func UpdateUser(c *fiber.Ctx) error {
	// Get the username from the URL params
	username := c.Params("username")

	// Parse the request body into a struct
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the updated time
	user.UpdatedAt = time.Now()

	// Update the user in the database
	filter := bson.M{"username": username}
	update := bson.M{"$set": user}
	if _, err := mongo.Users.UpdateOne(context.Background(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update user in database",
		})
	}

	// Return the updated user
	return c.JSON(user)
}

// DeleteUser deletes a user from the database by ID
func DeleteUser(c *fiber.Ctx) error {
	// Your implementation here
}

// ListUsers retrieves all users from the database
func ListUsers(c *fiber.Ctx) error {
	// Your implementation here
}

package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

// CreateUser inserts a new user into the database
func CreateUser(c *fiber.Ctx) error {
	// Get a handle to the users collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Parse the request body into a struct
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the created time
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Set the user's friends to an emtpy array
	user.ListOfFriends = []string{}

	// Insert the user into the database
	res, err := collection.InsertOne(context.Background(), user)
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
	// Get a handle to the users collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Get the user ID from the request parameters
	userID := c.Params("ID")

	// Convert the user ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Find the user in the database by ID
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
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

// UpdateUser updates a user in the database by ID
func UpdateUser(c *fiber.Ctx) error {
	// Get a handle to the users collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Get the username from the URL params
	userID := c.Params("ID")

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
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": user}
	if _, err := collection.UpdateOne(context.Background(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update user in database",
		})
	}

	// Return the updated user
	return c.JSON(user)
}

// DeleteUser deletes a user from the database by ID
func DeleteUser(c *fiber.Ctx) error {
	// Get a handle to the users collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Get the username from the URL parameters
	userID := c.Params("ID")

	// Delete the user from the database
	res, err := collection.DeleteOne(context.Background(), bson.M{"_id": userID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not delete user from database",
		})
	}

	// Check if a document was deleted
	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

// ListUsers retrieves all users from the database
func ListUsers(c *fiber.Ctx) error {
	// Get a handle to the users collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Find all users in the database
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve users from database",
		})
	}

	// Decode the cursor into a slice of users
	var users []models.User
	if err := cursor.All(context.Background(), &users); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode users from cursor",
		})
	}

	// Return the users
	return c.JSON(users)
}

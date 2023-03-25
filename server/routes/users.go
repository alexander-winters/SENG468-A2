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
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

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
	// Set the user's PostCount to 0
	user.PostCount = 0

	// Use a channel to send the result of the insert operation
	resultChan := make(chan *mongo.InsertOneResult)

	// Run the insert operation in a Go routine
	go func() {
		res, err := usersCollection.InsertOne(context.Background(), user)
		if err != nil {
			resultChan <- nil
			return
		}
		resultChan <- res
	}()

	// Wait for the insert operation to complete
	res := <-resultChan

	// Check if the insert operation was successful
	if res == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert user into database",
		})
	}

	// Set the ID of the user and return it
	user.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(user)
}

// GetUser retrieves a user from the database by username
func GetUser(c *fiber.Ctx) error {
	// Get a handle to the users collection
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username from the request parameters
	username := c.Params("username")

	// Find the user in the database by username
	var user models.User
	err := usersCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
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

// UpdateUser updates a user in the database by username
func UpdateUser(c *fiber.Ctx) error {
	// Get a handle to the users collection
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

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

	// Use a Go routine to update the user in the database
	updateChan := make(chan error)
	go func() {
		filter := bson.M{"username": username}
		update := bson.M{"$set": user}
		_, err := usersCollection.UpdateOne(context.Background(), filter, update)
		updateChan <- err
	}()

	// Wait for the update operation to complete
	err := <-updateChan
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update user in database",
		})
	}

	// Return the updated user
	return c.JSON(user)
}

// DeleteUser deletes a user from the database by username
func DeleteUser(c *fiber.Ctx) error {
	// Get a handle to the users collection
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username from the URL parameters
	username := c.Params("username")

	// Delete the user from the database
	res, err := usersCollection.DeleteOne(context.Background(), bson.M{"username": username})
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
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Find all users in the database
	cursor, err := usersCollection.Find(context.Background(), bson.M{})
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

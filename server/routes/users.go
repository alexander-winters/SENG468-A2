package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

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

	// Set the user's friends to an empty array
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

	// Serialize the user object to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not serialize user object",
		})
	}

	// Store the user object in Redis
	ctx := context.Background()
	err = rdb.Set(ctx, user.Username, userJSON, 0).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not store user in Redis",
		})
	}

	return c.JSON(user)
}

// GetUser retrieves a user from the database by username
func GetUser(c *fiber.Ctx) error {
	// Get a handle to the users collection
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username from the request parameters
	username := c.Params("username")

	// Check Redis cache for the user
	ctx := context.Background()
	userJSON, err := rdb.Get(ctx, username).Result()

	if err == nil {
		// User found in Redis cache
		var user models.User
		err := json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not deserialize user object",
			})
		}
		return c.JSON(user)
	} else if err != redis.Nil {
		// Redis error occurred
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve user from Redis",
		})
	}

	// User not found in Redis cache, query the database
	var user models.User
	err = usersCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
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

	// Store the user in Redis cache
	userJSONBytes, err := json.Marshal(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not serialize user object",
		})
	}
	userJSON = string(userJSONBytes)

	err = rdb.Set(ctx, user.Username, userJSON, 0).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not store user in Redis",
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

	// Check Redis cache
	ctx := context.Background()
	userJSON, err := rdb.Get(ctx, username).Result()
	if err != nil && err != redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not get user from Redis",
		})
	}

	if err == redis.Nil { // User not in Redis cache, update in the database
		filter := bson.M{"username": username}
		update := bson.M{"$set": user}
		_, err := usersCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not update user in database",
			})
		}
	} else { // User found in Redis cache, update and store it back
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not deserialize user object",
			})
		}

		// Update the user object
		user.UpdatedAt = time.Now()

		// Serialize the updated user object and store it back in Redis
		userJSONBytes, err := json.Marshal(user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not serialize user object",
			})
		}
		userJSON = string(userJSONBytes)
		err = rdb.Set(ctx, user.Username, userJSON, 0).Err()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not store user in Redis",
			})
		}
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

	// Use a Goroutine to delete the user from the database
	errChan := make(chan error)
	go func() {
		// Delete the user from the database
		res, err := usersCollection.DeleteOne(context.Background(), bson.M{"username": username})
		if err != nil {
			errChan <- err
			return
		}

		// Check if a document was deleted
		if res.DeletedCount == 0 {
			errChan <- fmt.Errorf("user not found")
			return
		}

		errChan <- nil
	}()

	// Wait for the Goroutine to complete or timeout
	select {
	case err := <-errChan:
		if err != nil {
			if err.Error() == "User not found" {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not delete user from database",
			})
		}

		return c.JSON(fiber.Map{
			"message": "User deleted successfully",
		})

	case <-time.After(5 * time.Second):
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Timed out while deleting user from database",
		})
	}
}

// ListUsers retrieves all users from the database
func ListUsers(c *fiber.Ctx) error {
	// Get a handle to the users collection
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Create a channel for receiving users
	usersChan := make(chan []models.User)

	// Retrieve users from the database in a goroutine
	go func() {
		// Find all users in the database
		cursor, err := usersCollection.Find(context.Background(), bson.M{})
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not retrieve users from database",
			})
			return
		}

		// Decode the cursor into a slice of users
		var users []models.User
		if err := cursor.All(context.Background(), &users); err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not decode users from cursor",
			})
			return
		}

		// Send the users to the channel
		usersChan <- users
	}()

	// Receive the users from the channel
	users := <-usersChan

	// Return the users
	return c.JSON(users)
}

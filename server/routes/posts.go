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

// CreatePost inserts a new post into the database
func CreatePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Parse the request body into a struct
	var post models.Post
	if err := c.BodyParser(&post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the created time
	post.CreatedAt = time.Now()

	// Insert the post into the database
	res, err := collection.InsertOne(context.Background(), post)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert post into database",
		})
	}

	// Set the ID of the post and return it
	post.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(post)
}

// GetPost retrieves a post from the database by ID
func GetPost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Get the post ID from the request parameters
	postID := c.Params("ID")

	// Convert the post ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	// Find the post in the database by ID
	var post models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Post not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve post from database",
		})
	}

	return c.JSON(post)
}

// UpdatePost updates a post in the database by ID
func UpdatePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Get the ID from the URL params
	postID := c.Params("ID")

	// Parse the request body into a struct
	var post models.Post
	if err := c.BodyParser(&post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the updated time
	post.UpdatedAt = time.Now()

	// Update the post in the database
	filter := bson.M{"_id": postID}
	update := bson.M{"$set": post}
	if _, err := collection.UpdateOne(context.Background(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update post in database",
		})
	}

	// Return the updated post
	return c.JSON(post)
}

// DeletePost deletes a post from the database by ID
func DeletePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Get the ID from the URL parameters
	postID := c.Params("ID")

	// Delete the post from the database
	res, err := collection.DeleteOne(context.Background(), bson.M{"_id": postID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not delete post from database",
		})
	}

	// Check if a document was deleted
	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Post deleted successfully",
	})
}

// ListPosts retrieves all posts from the database
func ListPosts(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Find all posts in the database
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve posts from database",
		})
	}

	// Decode the cursor into a slice of posts
	var posts []models.Post
	if err := cursor.All(context.Background(), &posts); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode posts from cursor",
		})
	}

	// Return the posts
	return c.JSON(posts)
}

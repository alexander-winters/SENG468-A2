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

// CreateComment inserts a new comment into the database
func CreateComment(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Parse the request body into a struct
	var comment models.Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the created time
	comment.CreatedAt = time.Now()

	// Insert the comment into the database
	res, err := collection.InsertOne(context.Background(), comment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert comment into database",
		})
	}

	// Set the ID of the comment and return it
	comment.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(comment)
}

// GetComment retrieves a comment from the database by ID
func GetComment(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Get the comment ID from the request parameters
	commentID := c.Params("ID")

	// Convert the comment ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Find the comment in the database by ID
	var comment models.Comment
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Comment not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve comment from database",
		})
	}

	return c.JSON(comment)
}

// UpdateComment updates a comment in the database by ID
func UpdateComment(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Get the ID from the URL params
	commentID := c.Params("ID")

	// Convert the comment ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Parse the request body into a struct
	var comment models.Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the updated time
	comment.UpdatedAt = time.Now()

	// Update the comment in the database
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": comment}
	if _, err := collection.UpdateOne(context.Background(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update comment in database",
		})
	}

	// Return the updated comment
	return c.JSON(comment)
}

// DeleteComment deletes a comment from the database by ID
func DeleteComment(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Get the ID from the URL parameters
	commentID := c.Params("ID")

	// Convert the comment ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Delete the post from the database
	res, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not delete comment from database",
		})
	}

	// Check if a document was deleted
	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Comment deleted successfully",
	})
}

// ListComments retrieves all comments from the database
func ListComments(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Find all comments in the database
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve users from database",
		})
	}

	// Decode the cursor into a slice of comments
	var comments []models.Comment
	if err := cursor.All(context.Background(), &comments); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode users from cursor",
		})
	}

	// Return the comments
	return c.JSON(comments)
}
package routes

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

// CreateComment inserts a new comment into the database for a specific post
func CreateComment(c *fiber.Ctx) error {
	// Get a handle to the comments and posts collections
	commentCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")
	postCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Parse the request body into a struct
	var comment models.Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the created time
	comment.CreatedAt = time.Now()

	// Get the post number from the request parameters
	postNum := c.Params("post_number")

	// Convert the post number to an integer
	postInt, err := strconv.Atoi(postNum)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Find the post in the database by post number
	var post models.Post
	err = postCollection.FindOne(context.Background(), bson.M{"postNum": postInt}).Decode(&post)
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

	// Add the comment to the post's comments array
	post.Comments = append(post.Comments, comment)

	// Update the post in the database
	_, err = postCollection.UpdateOne(context.Background(), bson.M{"postNum": postInt}, bson.M{"$set": bson.M{"comments": post.Comments}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update post in database",
		})
	}

	// Insert the comment into the database
	res, err := commentCollection.InsertOne(context.Background(), comment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert comment into database",
		})
	}

	// Set the ID of the comment and return it
	comment.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(comment)
}

// GetComment retrieves a comment from the database for a specific post by username and post number
func GetComment(c *fiber.Ctx) error {
	// Get a handle to the comments and posts collections
	commentCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")
	postCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Get the post number and username from the request parameters
	postNum := c.Params("post_number")
	username := c.Params("username")

	// Convert the post number to an integer
	postInt, err := strconv.Atoi(postNum)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Find the post in the database by post number and username
	var post models.Post
	err = postCollection.FindOne(context.Background(), bson.M{"postNum": postInt, "username": username}).Decode(&post)
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

	// Find the comment in the database by post ID
	var comment models.Comment
	err = commentCollection.FindOne(context.Background(), bson.M{"postID": post.ID}).Decode(&comment)
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

// GetComments retrieves all comments for a post by post number
func GetComments(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Get the post number from the request parameters
	postNum := c.Params("post_number")

	// Convert the post number to an integer
	postInt, err := strconv.Atoi(postNum)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Find all comments for the post in the database
	cursor, err := collection.Find(context.Background(), bson.M{"postNum": postInt})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve comments from database",
		})
	}

	// Decode the cursor into a slice of comments
	var comments []models.Comment
	if err := cursor.All(context.Background(), &comments); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode comments from cursor",
		})
	}

	// Return the comments
	return c.JSON(comments)
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

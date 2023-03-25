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
	// Get handles to the comments and posts collections
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")

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

	ctx := context.Background()
	var post models.Post
	done := make(chan bool)
	errChan := make(chan error)

	// Concurrently find the post in the database by post number
	go func() {
		err = postsCollection.FindOne(ctx, bson.M{"postNum": postInt}).Decode(&post)
		errChan <- err
		done <- true
	}()

	// Wait for the post to be found and handle any errors
	select {
	case err = <-errChan:
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
	case <-done:
	}

	// Add the comment to the post's comments array
	post.Comments = append(post.Comments, comment)

	// Concurrently update the post in the database
	go func() {
		_, err = postsCollection.UpdateOne(ctx, bson.M{"postNum": postInt}, bson.M{"$set": bson.M{"comments": post.Comments}})
		errChan <- err
		done <- true
	}()

	// Wait for the post to be updated and handle any errors
	select {
	case err = <-errChan:
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not update post in database",
			})
		}
	case <-done:
	}

	// Concurrently insert the comment into the database
	var res *mongo.InsertOneResult
	go func() {
		res, err = commentsCollection.InsertOne(ctx, comment)
		errChan <- err
		done <- true
	}()

	// Wait for the comment to be inserted and handle any errors
	select {
	case err = <-errChan:
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not insert comment into database",
			})
		}
	case <-done:
	}

	// Set the ID of the comment and return it
	comment.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(comment)
}

// GetComment retrieves a comment from the database for a specific post by username and post number
func GetComment(c *fiber.Ctx) error {
	// Get handles to the comments and posts collections
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")

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

	var post models.Post
	var comment models.Comment

	ctx := context.Background()
	errChan := make(chan error)
	done := make(chan bool)

	// Concurrently find the post in the database by post number and username
	go func() {
		err = postsCollection.FindOne(ctx, bson.M{"postNum": postInt, "username": username}).Decode(&post)
		errChan <- err
		done <- true
	}()

	// Concurrently find the comment in the database by post ID
	go func() {
		err = commentsCollection.FindOne(ctx, bson.M{"postID": post.ID}).Decode(&comment)
		errChan <- err
		done <- true
	}()

	for i := 0; i < 2; i++ {
		select {
		case err = <-errChan:
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error": "Post or comment not found",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Could not retrieve post or comment from database",
				})
			}
		case <-done:
		}
	}

	return c.JSON(comment)
}

// UpdateComment updates a comment in the database by ID
func UpdateComment(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")

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

	// Create channels for error handling and synchronization
	errChan := make(chan error)
	done := make(chan bool)

	ctx := context.Background()

	// Concurrently update the comment in the database
	go func() {
		filter := bson.M{"_id": objID}
		update := bson.M{"$set": comment}
		_, err := commentsCollection.UpdateOne(ctx, filter, update)
		errChan <- err
		done <- true
	}()

	// Wait for the comment to be updated and handle any errors
	select {
	case err = <-errChan:
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not update comment in database",
			})
		}
	case <-done:
	}

	// Return the updated comment
	return c.JSON(comment)
}

// DeleteComment deletes a comment from the database by ID
func DeleteComment(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")

	// Get the ID from the URL parameters
	commentID := c.Params("ID")

	// Convert the comment ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Create channels for error handling and synchronization
	errChan := make(chan error)
	done := make(chan bool)

	ctx := context.Background()

	// Concurrently delete the comment from the database
	go func() {
		res, err := commentsCollection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			errChan <- err
		} else if res.DeletedCount == 0 {
			errChan <- mongo.ErrNoDocuments
		} else {
			done <- true
		}
	}()

	// Wait for the comment to be deleted and handle any errors
	select {
	case err = <-errChan:
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Comment not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not delete comment from database",
		})
	case <-done:
	}

	return c.JSON(fiber.Map{
		"message": "Comment deleted successfully",
	})
}

// ListComments retrieves all comments for a post by post number
func ListComments(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")

	// Get the post number from the request parameters
	postNum := c.Params("post_number")

	// Convert the post number to an integer
	postInt, err := strconv.Atoi(postNum)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Create channels for error handling and synchronization
	errChan := make(chan error)
	commentsChan := make(chan []models.Comment)

	ctx := context.Background()

	// Concurrently find all comments for the post in the database
	go func() {
		cursor, err := commentsCollection.Find(ctx, bson.M{"postNum": postInt})
		if err != nil {
			errChan <- err
			return
		}

		// Decode the cursor into a slice of comments
		var comments []models.Comment
		if err := cursor.All(ctx, &comments); err != nil {
			errChan <- err
			return
		}

		commentsChan <- comments
	}()

	// Wait for the comments to be retrieved and handle any errors
	var comments []models.Comment
	select {
	case err = <-errChan:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve comments from database",
		})
	case comments = <-commentsChan:
	}

	// Return the comments
	return c.JSON(comments)
}

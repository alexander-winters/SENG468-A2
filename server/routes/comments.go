package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/alexander-winters/SENG468-A2/kafka-docker/kafkaService"
	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

var kafkaBrokerURL = "kafka:9092"

// CreateComment inserts a new comment into the database for a specific post
func CreateComment(c *fiber.Ctx) error {
	// Get handles to the comments and posts collections
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")

	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("post_number"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Retrieve the post by username and postNumber
	post, err := GetPostByUsername(username, postNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve user",
		})
	}

	// Parse the request body into a struct
	var comment models.Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the created time
	comment.CreatedAt = time.Now()

	// Add the comment to the post's comments array
	post.Comments = append(post.Comments, comment)

	// Update the comment in the database
	res, err := commentsCollection.InsertOne(c.Context(), comment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert comment into database",
		})
	}

	// Update the post in the database
	filter := bson.M{"username": username, "post_number": postNumber}
	update := bson.M{"$set": bson.M{"comments": post.Comments}}
	if _, err := postsCollection.UpdateOne(c.Context(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update post in database",
		})
	}

	// Update the post in the Redis cache
	postJSON, err := json.Marshal(post)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not serialize post to Redis cache",
		})
	}
	if err := rdb.Set(c.Context(), "post:"+username+":"+strconv.Itoa(postNumber), postJSON, 0).Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not set post in Redis cache",
		})
	}

	// Set the ID of the comment and return it
	comment.ID = res.InsertedID.(primitive.ObjectID)

	// Create a Notification
	notification := models.Notification{
		UserID:     comment.UserID,
		Username:   username,
		Type:       models.CommentCreatedNotification,
		PostID:     comment.PostID,
		CommentID:  comment.ID,
		Recipient:  post.Username,
		Content:    comment.Content,
		ReadStatus: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Initialize Kafka producer and consumer
	kafkaProducer := kafkaService.CreateKafkaProducer(kafkaBrokerURL)
	kafkaConsumer := kafkaService.CreateKafkaConsumer(kafkaBrokerURL, "comment")

	ks := kafkaService.NewKafkaService(kafkaProducer, kafkaConsumer)

	// Send a notification to the post owner
	err = ks.SendUserNotification(notification)
	if err != nil {
		log.Printf("Could not send notification: %v", err)
	}
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

	ctx := c.Context()
	errChan := make(chan error)
	done := make(chan bool)

	// Concurrently find the post in the Redis cache or the database by post number and username
	go func() {
		cacheKey := "post:" + username + ":" + strconv.Itoa(postInt)
		if err := rdb.Get(ctx, cacheKey).Scan(&post); err == redis.Nil {
			err = postsCollection.FindOne(ctx, bson.M{"postNum": postInt, "username": username}).Decode(&post)
			if err == nil {
				postJSON, _ := json.Marshal(post)
				rdb.Set(ctx, cacheKey, postJSON, 0)
			}
		}
		errChan <- err
		done <- true
	}()

	// Concurrently find the comment in the Redis cache or the database by post ID
	go func() {
		cacheKey := "comment:" + post.ID.Hex()
		if err := rdb.Get(ctx, cacheKey).Scan(&comment); err == redis.Nil {
			err = commentsCollection.FindOne(ctx, bson.M{"postID": post.ID}).Decode(&comment)
			if err == nil {
				commentJSON, _ := json.Marshal(comment)
				rdb.Set(ctx, cacheKey, commentJSON, 0)
			}
		}
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
					"error": "Could not retrieve post or comment from database or cache",
				})
			}
		case <-done:
		}
	}

	return c.JSON(comment)
}

// UpdateComment updates a comment in the database by username and post number
func UpdateComment(c *fiber.Ctx) error {
	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Retrieve the post
	post, err := GetPostByUsername(username, postNumber)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Get a handle to the comments collection
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")

	// Parse the request body into a struct
	var updatedComment models.Comment
	if err := c.BodyParser(&updatedComment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the updated time
	updatedComment.UpdatedAt = time.Now()

	// Find the comment with the given ID
	var existingComment *models.Comment
	for _, c := range post.Comments {
		if c.ID == updatedComment.ID {
			existingComment = &c
			break
		}
	}

	// Check if the comment was found
	if existingComment == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	// Update the comment in the database
	filter := bson.M{"_id": existingComment.ID}
	update := bson.M{"$set": bson.M{"content": updatedComment.Content, "updated_at": updatedComment.UpdatedAt}}
	_, err = commentsCollection.UpdateOne(c.Context(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update comment in database",
		})
	}

	// Return the updated comment
	return c.JSON(updatedComment)
}

// DeleteComment deletes a comment from the database by username and post number
func DeleteComment(c *fiber.Ctx) error {
	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Retrieve the post
	post, err := GetPostByUsername(username, postNumber)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Get a handle to the comments collection
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")

	// Get the comment ID from the request body
	var commentToDelete models.Comment
	if err := c.BodyParser(&commentToDelete); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Find the comment with the given ID
	var existingComment *models.Comment
	for _, c := range post.Comments {
		if c.ID == commentToDelete.ID {
			existingComment = &c
			break
		}
	}

	// Check if the comment was found
	if existingComment == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	// Delete the comment from the database
	res, err := commentsCollection.DeleteOne(c.Context(), bson.M{"_id": existingComment.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not delete comment from database",
		})
	} else if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found",
		})
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
	case err := <-errChan:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Could not process request: %v", err),
		})
	case comments = <-commentsChan:
	}

	// Return the comments
	return c.JSON(comments)
}

func LikeComment(c *fiber.Ctx) error {
	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Get a handle to the comments collection
	commentsCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Retrieve the existing post
	existingComment, err := GetPostByUsername(username, postNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Post not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve post",
		})
	}

	// Check if the user has already liked the post
	liked := false
	for _, like := range existingComment.Likes {
		if like.Username == c.Locals("username").(string) {
			liked = true
			break
		}
	}

	// Add the like and increment the likes count if the user has not already liked the post
	if !liked {
		like := models.Like{
			Username: c.Locals("username").(string),
			LikedAt:  time.Now(),
		}
		existingComment.Likes = append(existingComment.Likes, like)
		existingComment.NumberOfLikes = existingComment.NumberOfLikes + 1
	}
	// Create a Notification
	notification := models.Notification{
		UserID:     existingComment.UserID,
		Username:   username,
		Type:       models.PostCreatedNotification,
		PostID:     existingComment.ID,
		Recipient:  existingComment.Username,
		Content:    existingComment.Content,
		ReadStatus: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Initialize Kafka producer and consumer
	kafkaProducer := kafkaService.CreateKafkaProducer(kafkaBrokerURL)
	kafkaConsumer := kafkaService.CreateKafkaConsumer(kafkaBrokerURL, "like")

	ks := kafkaService.NewKafkaService(kafkaProducer, kafkaConsumer)

	// Send a notification to the post owner
	err = ks.SendUserNotification(notification)
	if err != nil {
		log.Printf("Could not send notification: %v", err)
	}

	// Update the post in the database
	filter := bson.M{"username": username, "post_number": postNumber}
	update := bson.M{"$set": bson.M{"likes": existingComment.Likes, "likes_count": existingComment.NumberOfLikes}}
	_, err = commentsCollection.UpdateOne(c.Context(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update post in database",
		})
	}

	// Update the post in Redis cache
	postKey := fmt.Sprintf("post:%s:%d", username, postNumber)
	postJSONBytes, err := json.Marshal(existingComment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not serialize post object",
		})
	}
	postJSON := string(postJSONBytes)

	err = rdb.Set(c.Context(), postKey, postJSON, 0).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not store post in Redis",
		})
	}

	// Return the updated post
	return c.JSON(existingComment)
}

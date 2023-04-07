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

// CreatePost inserts a new post into the database
func CreatePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username from the URL parameters
	username := c.Params("username")

	// Retrieve the user by username
	user, err := GetUserByUsername(username)
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
	var post models.Post
	if err := c.BodyParser(&post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the post number, username, and created time
	post.PostNumber = user.PostCount + 1
	post.Username = user.Username
	post.Comments = []models.Comment{}
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	// Insert the post into the database
	res, err := postsCollection.InsertOne(c.Context(), post)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert post into database",
		})
	}

	// Increment the user's post count and update the user in the database
	user.PostCount++
	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"postCount": user.PostCount}}

	_, err = usersCollection.UpdateOne(c.Context(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update user in database",
		})
	}

	// Update the user in Redis cache
	userJSONBytes, err := json.Marshal(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not serialize user object",
		})
	}
	userJSON := string(userJSONBytes)

	err = rdb.Set(c.Context(), "user:"+user.Username, userJSON, 0).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not store user in Redis",
		})
	}

	// Set the ID of the comment and return it
	post.ID = res.InsertedID.(primitive.ObjectID)

	// Get the friends of the user who created the post
	friends, err := GetUserFriends(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve user friends",
		})
	}

	// Iterate through the friends and send notifications
	for _, friend := range friends {
		// Create a Notification
		notification := models.Notification{
			UserID:     post.UserID,
			Username:   username,
			Type:       models.PostCreatedNotification,
			PostID:     post.ID,
			Recipient:  friend,
			Content:    post.Content,
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
	}

	// Return the created post
	return c.JSON(post)
}

// GetPostbyUsername retrieves a post by username and postNumber, first checking Redis cache, then the database
func GetPostByUsername(username string, postNumber int) (*models.Post, error) {
	// Get a handle to the posts collection
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")

	// Check Redis cache for the post
	ctx := context.Background()
	postKey := fmt.Sprintf("post:%s:%d", username, postNumber)
	postJSON, err := rdb.Get(ctx, postKey).Result()

	if err == redis.Nil {
		// Post not found in Redis cache, query the database
		var post models.Post
		err = postsCollection.FindOne(ctx, bson.M{"username": username, "post_number": postNumber}).Decode(&post)
		if err != nil {
			return nil, err
		}

		// Store the post in Redis cache
		postJSONBytes, err := json.Marshal(post)
		if err != nil {
			return nil, err
		}
		postJSON = string(postJSONBytes)

		err = rdb.Set(ctx, postKey, postJSON, 0).Err()
		if err != nil {
			return nil, err
		}

		return &post, nil
	} else if err != nil {
		// Redis error occurred
		return nil, err
	} else {
		// Post found in Redis cache
		var post models.Post
		err := json.Unmarshal([]byte(postJSON), &post)
		if err != nil {
			return nil, err
		}
		return &post, nil
	}

}

// GetPost retrieves a post from the database by username and post number
func GetPost(c *fiber.Ctx) error {
	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("post_number"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Retreive the post
	post, err := GetPostByUsername(username, postNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Post not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retreive post",
		})
	}

	return c.JSON(post)
}

// UpdatePost updates a post in the database by username and post number
func UpdatePost(c *fiber.Ctx) error {
	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Parse the request body into a struct
	var updatedPost models.Post
	if err := c.BodyParser(&updatedPost); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the updated time
	updatedPost.UpdatedAt = time.Now()

	// Get a handle to the posts collection
	postsCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Retrieve the existing post
	existingPost, err := GetPostByUsername(username, postNumber)
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
	for _, like := range existingPost.Likes {
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
		updatedPost.Likes = append(existingPost.Likes, like)
		updatedPost.NumberOfLikes = existingPost.NumberOfLikes + 1
	}
	// Create a Notification
	notification := models.Notification{
		UserID:     updatedPost.UserID,
		Username:   username,
		Type:       models.PostCreatedNotification,
		PostID:     updatedPost.ID,
		Recipient:  updatedPost.Username,
		Content:    updatedPost.Content,
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
	update := bson.M{"$set": bson.M{"content": updatedPost.Content, "updated_at": updatedPost.UpdatedAt, "likes": updatedPost.Likes, "likes_count": updatedPost.NumberOfLikes}}
	_, err = postsCollection.UpdateOne(c.Context(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update post in database",
		})
	}

	// Update the post in Redis cache
	postKey := fmt.Sprintf("post:%s:%d", username, postNumber)
	postJSONBytes, err := json.Marshal(updatedPost)
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
	return c.JSON(updatedPost)
}

// DeletePost deletes a post from the database by username and post number
func DeletePost(c *fiber.Ctx) error {
	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("post_number"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Get handles to the posts and users collections in the database
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Delete the post from the database
	res, err := postsCollection.DeleteOne(c.Context(), bson.M{"username": username, "post_number": postNumber})
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

	// Retrieve the user and decrement the user's post count
	user, err := GetUserByUsername(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve user",
		})
	}
	user.PostCount--

	// Update the user in the database
	filter := bson.M{"username": username}
	update := bson.M{"$set": user}
	if _, err := usersCollection.UpdateOne(c.Context(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update user in database",
		})
	}

	// Remove the post from Redis cache
	postKey := fmt.Sprintf("post:%s:%d", username, postNumber)
	err = rdb.Del(c.Context(), postKey).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not remove post from Redis",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Post deleted successfully",
	})
}

func GetUserPosts(ctx context.Context, username string) ([]models.Post, error) {
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")
	cursor, err := postsCollection.Find(ctx, bson.M{"username": username})
	if err != nil {
		return nil, err
	}

	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

// ListUserPosts retrieves all posts of a single user from the database by username
func ListUserPosts(c *fiber.Ctx) error {
	// Get the username from the URL parameters
	username := c.Params("username")

	posts, err := GetUserPosts(c.Context(), username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve posts from database",
		})
	}

	// Return the posts
	return c.JSON(posts)
}

// ListPosts retrieves all posts from the database
func ListAllPosts(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")

	// Find all posts in the database using a cursor
	ctx := c.Context()
	cursor, err := postsCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve posts from database",
		})
	}

	// Decode the cursor into a slice of posts
	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode posts from cursor",
		})
	}

	// Check for errors during iteration
	if err := cursor.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error during iteration",
		})
	}

	// Return the posts
	return c.JSON(posts)
}

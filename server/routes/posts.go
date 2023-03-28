package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

// CreatePost inserts a new post into the database
func CreatePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")

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
	_, err = postsCollection.InsertOne(c.Context(), post)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert post into database",
		})
	}

	// Increment the user's post count
	user.PostCount++

	// Update the user in the database and Redis cache
	err = UpdateUser(c.Context(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update user",
		})
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
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
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

	// Update the post in the database
	filter := bson.M{"username": username, "post_number": postNumber}
	update := bson.M{"$set": updatedPost}
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
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
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

// ListUserPosts retrieves all posts of a single user from the database by username
func ListUserPosts(c *fiber.Ctx) error {
	// Get a handle to the posts and users collections
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username from the URL parameters
	username := c.Params("username")

	// Find the user in the database by username
	var user models.User
	userChan := make(chan error)
	go func() {
		err := usersCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
		userChan <- err
	}()

	// Find all posts in the database for the specified user
	cursor, err := postsCollection.Find(context.Background(), bson.M{"username": username})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve posts from database",
		})
	}

	// Decode the cursor into a slice of posts
	var posts []models.Post
	postsChan := make(chan error)
	go func() {
		if err := cursor.All(context.Background(), &posts); err != nil {
			postsChan <- err
		}
		postsChan <- nil
	}()

	// Check for any errors from the concurrent calls
	var userErr, postsErr error
	for i := 0; i < 2; i++ {
		select {
		case err := <-userChan:
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error": "User not found",
					})
				}
				userErr = err
			}
		case err := <-postsChan:
			if err != nil {
				postsErr = err
			}
		}
	}
	if userErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve user from database",
		})
	}
	if postsErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode posts from cursor",
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
	ctx := context.Background()
	cursor, err := postsCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve posts from database",
		})
	}

	// Define a channel to receive the decoded posts
	postChan := make(chan models.Post)
	defer close(postChan)

	// Define a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Iterate over the cursor and decode each post into a channel
	for cursor.Next(ctx) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var post models.Post
			if err := cursor.Decode(&post); err != nil {
				log.Printf("Error decoding post: %s", err)
				return
			}
			postChan <- post
		}()
	}

	// Wait for all goroutines to finish decoding posts
	go func() {
		wg.Wait()
		close(postChan)
	}()

	// Collect all decoded posts into a slice
	var posts []models.Post
	for post := range postChan {
		posts = append(posts, post)
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

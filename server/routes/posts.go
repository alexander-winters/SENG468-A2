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
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// GetPost retrieves a post from the database by username and post number
func GetPost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")

	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Check Redis cache for the post
	ctx := c.Context()
	postKey := fmt.Sprintf("post:%s:%d", username, postNumber)
	postJSON, err := rdb.Get(ctx, postKey).Result()

	var post models.Post
	if err == redis.Nil {
		// Post not found in Redis cache, query the database
		err = postsCollection.FindOne(ctx, bson.M{"username": username, "post_number": postNumber}).Decode(&post)
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

		// Store the post in Redis cache
		postJSONBytes, err := json.Marshal(post)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not serialize post object",
			})
		}
		postJSON = string(postJSONBytes)

		err = rdb.Set(ctx, postKey, postJSON, 0).Err()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not store post in Redis",
			})
		}
	} else if err != nil {
		// Redis error occurred
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve post from Redis",
		})
	} else {
		// Post found in Redis cache
		err := json.Unmarshal([]byte(postJSON), &post)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not deserialize post object",
			})
		}
	}

	return c.JSON(post)
}

// UpdatePost updates a post in the database by ID
func UpdatePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postsCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")
	usersCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")
	// Get the ID from the URL params
	postID := c.Params("ID")

	// Get the username from the URL parameters
	username := c.Params("username")

	ctx := context.Background()
	// Try to get the user from Redis cache
	userJSON, err := rdb.Get(ctx, "user:"+username).Result()

	// Create a channel to receive the user data
	userChan := make(chan *models.User)

	if err == redis.Nil {
		// User not found in Redis cache, find the user in the database by username
		go func() {
			var user models.User
			err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error": "User not found",
					})
				}
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Could not retrieve user from database",
				})
				return
			}
			userChan <- &user
		}()
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve user from Redis cache",
		})
	} else {
		// Deserialize the user from Redis cache and send it to the channel
		go func() {
			var user models.User
			if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Could not deserialize user from Redis cache",
				})
				return
			}
			userChan <- &user
		}()
	}

	// Convert the post ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	// Parse the request body into a struct
	var post models.Post
	if err := c.BodyParser(&post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Wait for the user data to be received from the channel
	user := <-userChan

	// Check if the post belongs to the user
	if post.Username != user.Username {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Post does not belong to user",
		})
	}

	// Set the updated time
	post.UpdatedAt = time.Now()

	// Create a channel to receive the result of the post update
	postChan := make(chan error)
	// Update the post in the database
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": post}
	go func() {
		_, err := postsCollection.UpdateOne(ctx, filter, update)
		postChan <- err
	}()

	// Wait for the post update to complete
	err = <-postChan
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update post in database",
		})
	}

	// Return the updated post
	return c.JSON(post)
}

// DeletePost deletes a post from the database by username and post number
func DeletePost(c *fiber.Ctx) error {
	// Get handles to the posts and users collections in the database
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
	if err != nil {
		// Return an error message if the post number is invalid
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	var user models.User
	var post models.Post

	// Set up a background context, error channel and done channel for the goroutines
	ctx := context.Background()
	errChan := make(chan error)
	done := make(chan bool)

	// Try to get the user from Redis cache
	userJSON, err := rdb.Get(ctx, "user:"+username).Result()

	if err == redis.Nil {
		// Find the user in the database using a goroutine
		go func() {
			err = usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
			errChan <- err
			done <- true
		}()
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve user from Redis cache",
		})
	} else {
		// Deserialize the user from Redis cache
		if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not deserialize user from Redis cache",
			})
		}
		done <- true
	}

	// Find the post in the database using a goroutine
	go func() {
		err = postsCollection.FindOne(ctx, bson.M{"username": username, "post_number": postNumber}).Decode(&post)
		errChan <- err
		done <- true
	}()

	// Wait for both goroutines to complete or for an error to occur
	for i := 0; i < 2; i++ {
		select {
		case err = <-errChan:
			// Handle any errors that occurred during the goroutines
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error": "User or post not found",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Could not retrieve user or post from database",
				})
			}
		case <-done:
		}
	}

	// Delete the post from the database
	res, err := postsCollection.DeleteOne(ctx, bson.M{"_id": post.ID})
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

	// Decrement the user's post count and update the user in the database
	user.PostCount--
	filter := bson.M{"username": username}
	update := bson.M{"$set": user}
	if _, err := usersCollection.UpdateOne(ctx, filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update user in database",
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

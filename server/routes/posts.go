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

// CreatePost inserts a new post into the database
func CreatePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")
	userCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Get the username from the URL parameters
	username := c.Params("username")

	// Create a channel to receive the user data
	userChan := make(chan *models.User)
	// Find the user in the database by username
	go func() {
		var user models.User
		err := userCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
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

	// Parse the request body into a struct
	var post models.Post
	if err := c.BodyParser(&post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Wait for the user data to be received from the channel
	user := <-userChan

	// Set the post number, username, and created time
	post.PostNumber = user.PostCount + 1
	post.Username = user.Username
	// Initialize an empty array of comments
	post.Comments = []models.Comment{}
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	// Create a channel to receive the result of the post insert
	postChan := make(chan error)
	// Insert the post into the database
	go func() {
		_, err := postCollection.InsertOne(context.Background(), post)
		postChan <- err
	}()

	// Increment the user's post count and update the user in the database
	user.PostCount++
	filter := bson.M{"username": username}
	update := bson.M{"$set": user}

	// Create a channel to receive the result of the user update
	userChan2 := make(chan error)
	// Update the user in the database
	go func() {
		_, err := userCollection.UpdateOne(context.Background(), filter, update)
		userChan2 <- err
	}()

	// Wait for the post insert and user update to complete
	err1, err2 := <-postChan, <-userChan2
	if err1 != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert post into database",
		})
	}
	if err2 != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update user in database",
		})
	}

	// Return the created post
	return c.JSON(post)
}

// GetPost retrieves a post from the database by username and post number
func GetPost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Get the username and post number from the request parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Create a channel to receive the post data
	postChan := make(chan *models.Post)
	// Find the post in the database by username and post number
	go func() {
		var post models.Post
		err = collection.FindOne(context.Background(), bson.M{"username": username, "post_number": postNumber}).Decode(&post)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Post not found",
				})
			}
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not retrieve post from database",
			})
			return
		}
		postChan <- &post
	}()

	// Wait for the post data to be received from the channel
	post := <-postChan

	return c.JSON(post)
}

// UpdatePost updates a post in the database by ID
func UpdatePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")
	userCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")
	// Get the ID from the URL params
	postID := c.Params("ID")

	// Get the username from the URL parameters
	username := c.Params("username")

	// Find the user in the database by username
	var user models.User
	err := userCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
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

	// Check if the post belongs to the user
	if post.Username != user.Username {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Post does not belong to user",
		})
	}

	// Set the updated time
	post.UpdatedAt = time.Now()

	// Update the post in the database
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": post}
	if _, err := postCollection.UpdateOne(context.Background(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update post in database",
		})
	}

	// Return the updated post
	return c.JSON(post)
}

// DeletePost deletes a post from the database by username and post number
func DeletePost(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")
	userCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Get the username and post number from the URL parameters
	username := c.Params("username")
	postNumber, err := strconv.Atoi(c.Params("postNumber"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post number",
		})
	}

	// Find the user in the database by username
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
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

	// Find the post in the database by username and post number
	var post models.Post
	err = postCollection.FindOne(context.Background(), bson.M{"username": username, "post_number": postNumber}).Decode(&post)
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

	// Delete the post from the database
	res, err := postCollection.DeleteOne(context.Background(), bson.M{"_id": post.ID})
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
	if _, err := userCollection.UpdateOne(context.Background(), filter, update); err != nil {
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
	postCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")
	userCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Get the username from the URL parameters
	username := c.Params("username")

	// Find the user in the database by username
	var user models.User
	err := userCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
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

	// Find all posts in the database for the specified user
	cursor, err := postCollection.Find(context.Background(), bson.M{"username": user.Username})
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

// ListPosts retrieves all posts from the database
func ListAllPosts(c *fiber.Ctx) error {
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

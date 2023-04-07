package routes

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

// PostReport retrieves a report of all posts created by a given user
func PostReport(c *fiber.Ctx) error {
	// Get a handle to the posts and users collections
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username from the request parameters
	username := c.Params("username")

	// Create a context
	ctx := context.Background()

	// Create channels for error handling and synchronization
	errChan := make(chan error, 1)
	userChan := make(chan models.User, 1)

	// Concurrently find the user document that matches the username
	go func() {
		var user models.User
		err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- user
	}()

	// Wait for the concurrent task to finish and handle any errors
	var userID primitive.ObjectID
	select {
	case err := <-errChan:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid username: %v", err),
		})
	case user := <-userChan:
		userID = user.ID
	}

	// Create the pipeline for the aggregation query
	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"user_id": userID,
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": "$user_id",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"user_id": "$_id",
				"count":   1,
				"_id":     0,
			},
		},
	}

	// Create a channel for aggregation results
	resultsChan := make(chan []models.PostReport, 1)

	// Concurrently execute the aggregation query
	go func() {
		cursor, err := postsCollection.Aggregate(ctx, pipeline)
		if err != nil {
			errChan <- err
			return
		}

		// Decode the results into a slice of PostReport objects
		var results []models.PostReport
		if err := cursor.All(ctx, &results); err != nil {
			errChan <- err
			return
		}

		// Send the results to the results channel
		resultsChan <- results
	}()

	// Wait for the concurrent task to finish and handle any errors
	select {
	case err := <-errChan:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Could not generate report: %v", err),
		})
	case results := <-resultsChan:
		// If there are no results, create an empty PostReport
		if len(results) == 0 {
			results = append(results, models.PostReport{UserID: userID, Username: username, Count: 0})
		}

		// Return the report
		return c.JSON(results)
	}
}

// UserCommentReport retrieves a report of comments created by user
func UserCommentReport(c *fiber.Ctx) error {
	// Get a handle to the comments and users collection
	commentsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("comments")
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username from the request parameters
	username := c.Params("username")

	// Create a context
	ctx := context.Background()

	// Create channels for error handling and synchronization
	errChan := make(chan error, 1)
	userChan := make(chan models.User, 1)

	// Concurrently find the user in the users collection
	go func() {
		var user models.User
		err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- user
	}()

	// Wait for the concurrent task to finish and handle any errors
	var userID primitive.ObjectID
	select {
	case err := <-errChan:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("User not found: %v", err),
		})
	case user := <-userChan:
		userID = user.ID
	}

	// Create channels for comment cursor and comment results
	commentsChan := make(chan []models.Comment, 1)

	// Concurrently find all comments created by the user and store them in a slice
	go func() {
		cursor, err := commentsCollection.Find(ctx, bson.M{"user_id": userID})
		if err != nil {
			errChan <- err
			return
		}

		var comments []models.Comment
		for cursor.Next(ctx) {
			var comment models.Comment
			if err := cursor.Decode(&comment); err != nil {
				errChan <- err
				return
			}
			comments = append(comments, comment)
		}

		commentsChan <- comments
	}()

	// Wait for the concurrent task to finish and handle any errors
	var comments []models.Comment
	select {
	case err := <-errChan:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Could not retrieve comments from database: %v", err),
		})
	case comments = <-commentsChan:
	}

	// Return the JSON output directly
	return c.JSON(fiber.Map{
		"username":       username,
		"total_comments": len(comments),
		"comments":       comments,
	})
}

// LikeReport retrieves a report on likes given or received by a user
func LikeReport(c *fiber.Ctx) error {
	// Get a handle to the likes and users collections
	likesCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("likes")
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the username from the URL params
	username := c.Params("username")

	// Create a context
	ctx := context.Background()

	// Create channels for error handling and synchronization
	errChan := make(chan error, 1)
	userChan := make(chan models.User, 1)

	// Concurrently find the user in the users collection
	go func() {
		var user models.User
		err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- user
	}()

	// Wait for the concurrent task to finish and handle any errors
	var userID primitive.ObjectID
	select {
	case err := <-errChan:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid username: %v", err),
		})
	case user := <-userChan:
		userID = user.ID
	}

	// Create a pipeline to retrieve the total number of likes given and received by the user
	pipeline := bson.A{
		bson.M{"$match": bson.M{"user_id": userID}},
		bson.M{"$group": bson.M{
			"_id":         "$user_id",
			"likes_given": bson.M{"$sum": 1},
			"likes_received": bson.M{"$sum": bson.M{"$cond": bson.A{
				bson.M{"$eq": bson.A{"$liked_by_id", userID}},
				1,
				0,
			}}},
		}},
	}

	// Create channels for cursor and report results
	reportsChan := make(chan []models.LikeReport, 1)

	// Concurrently execute the pipeline and retrieve the results
	go func() {
		cursor, err := likesCollection.Aggregate(ctx, pipeline)
		if err != nil {
			errChan <- err
			return
		}

		var reports []models.LikeReport
		if err := cursor.All(ctx, &reports); err != nil {
			errChan <- err
			return
		}

		reportsChan <- reports
	}()

	// Wait for the concurrent task to finish and handle any errors
	var reports []models.LikeReport
	select {
	case err := <-errChan:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Could not retrieve likes from database: %v", err),
		})
	case reports = <-reportsChan:
		// If there are no results, create an empty LikeReport
		if len(reports) == 0 {
			reports = append(reports, models.LikeReport{UserID: userID, Username: username, LikesGiven: 0, LikesReceived: 0})
		}
	}

	// Return the reports
	return c.JSON(reports)
}

package routes

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

// PostReport retrieves a report of all posts created by a given user
func PostReport(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	postCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Get the user ID from the request parameters
	userID := c.Params("user_id")

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

	// Execute the aggregation query
	cursor, err := postCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not generate report",
		})
	}

	// Decode the results into a slice of PostReport objects
	var results []models.PostReport
	if err := cursor.All(context.Background(), &results); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode report results",
		})
	}

	// Return the report
	return c.JSON(results)
}

// UserCommentReport retrieves a report of comments created by a user
func UserCommentReport(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	commentCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Get the username from the request parameters
	username := c.Params("username")

	// Find the user ID from the users collection
	userCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")
	var user models.User
	if err := userCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Find all comments created by the user
	cursor, err := commentCollection.Find(context.Background(), bson.M{"user_id": user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve comments from database",
		})
	}

	// Calculate the total number of comments
	totalComments := 0
	for cursor.Next(context.Background()) {
		totalComments++
	}

	// Create the comment reports object
	commentReports := models.CommentReport{
		UserID:        user.ID,
		Username:      user.Username,
		TotalComments: totalComments,
	}

	return c.JSON(commentReports)
}

// LikeReport retrieves a report on likes given or received by a user
func LikeReport(c *fiber.Ctx) error {
	// Get a handle to the likes and users collections
	likesCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("likes")
	usersCollection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("users")

	// Get the username from the URL params
	username := c.Params("username")

	// Find the user document that matches the username and extract the ID
	var user models.User
	err := usersCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid username",
		})
	}
	userID := user.ID

	// Create a pipeline to retrieve the total number of likes given and received by the user
	pipeline := bson.A{
		bson.M{"$match": bson.M{"user_id": userID}},
		bson.M{"$group": bson.M{
			"_id":         "$user_id",
			"likes_given": bson.M{"$sum": 1},
			"likes_received": bson.M{"$sum": bson.M{"$cond": bson.A{
				bson.M{"$eq": bson.A{"$liked_by", userID}},
				1,
				0,
			}}},
		}},
	}

	// Execute the pipeline and retrieve the results
	cursor, err := likesCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve likes from database",
		})
	}

	// Decode the cursor into a slice of LikeReports
	var reports []models.LikeReport
	if err := cursor.All(context.Background(), &reports); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode likes from cursor",
		})
	}

	// Return the reports
	return c.JSON(reports)
}

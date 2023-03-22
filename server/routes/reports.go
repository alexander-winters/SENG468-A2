package routes

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

func PostReports(c *fiber.Ctx) error {
	// Get a handle to the posts collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("posts")

	// Create the pipeline for the aggregation query
	pipeline := bson.A{
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
	cursor, err := collection.Aggregate(context.Background(), pipeline)
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

// GetUserCommentReports retrieves a report of comments created by a user
func GetUserCommentReports(c *fiber.Ctx) error {
	// Get a handle to the comments collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("comments")

	// Get the user ID from the request parameters
	userID := c.Params("id")

	// Convert the user ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Find all comments created by the user
	cursor, err := collection.Find(context.Background(), bson.M{"user_id": objID})
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
	commentReports := models.CommentReports{
		UserID:        objID,
		TotalComments: totalComments,
	}

	return c.JSON(commentReports)
}

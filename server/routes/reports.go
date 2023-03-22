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

// UserCommentReports retrieves a report of comments created by a user
func UserCommentReports(c *fiber.Ctx) error {
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

// LikeReports retrieves a report on likes given or received by a user
func LikeReports(c *fiber.Ctx) error {
	// Get a handle to the likes collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("likes")

	// Get the user ID from the URL params
	userID := c.Params("id")

	// Convert the user ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Create a pipeline to retrieve the total number of likes given and received by the user
	pipeline := bson.A{
		bson.M{"$match": bson.M{"user_id": objID}},
		bson.M{"$group": bson.M{
			"_id":         "$user_id",
			"likes_given": bson.M{"$sum": 1},
			"likes_received": bson.M{"$sum": bson.M{"$cond": bson.A{
				bson.M{"$eq": bson.A{"$liked_by", objID}},
				1,
				0,
			}}},
		}},
	}

	// Execute the pipeline and retrieve the results
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve likes from database",
		})
	}

	// Decode the cursor into a slice of LikeReports
	var reports []models.LikeReports
	if err := cursor.All(context.Background(), &reports); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode likes from cursor",
		})
	}

	// Return the reports
	return c.JSON(reports)
}

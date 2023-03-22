package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

// CreateNotification inserts a new notification into the database
func CreateNotification(c *fiber.Ctx) error {
	// Get a handle to the notifications collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("notifications")

	// Parse the request body into a struct
	var notification models.Notification
	if err := c.BodyParser(&notification); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the created time
	notification.CreatedAt = time.Now()

	// Insert the notification into the database
	res, err := collection.InsertOne(context.Background(), notification)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not insert notification into database",
		})
	}

	// Set the ID of the notification and return it
	notification.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(notification)
}

// MarkNotificationAsRead updates the read status of a notification in the database by ID
func MarkNotificationAsRead(c *fiber.Ctx) error {
	// Get a handle to the notifications collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("notifications")

	// Get the ID from the URL params
	notificationID := c.Params("ID")

	// Convert the notification ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid notification ID",
		})
	}

	// Update the notification in the database
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"read_status": true, "updated_at": time.Now()}}
	if _, err := collection.UpdateOne(context.Background(), filter, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not update notification in database",
		})
	}

	// Return a success message
	return c.JSON(fiber.Map{
		"message": "Notification read status updated successfully",
	})
}

// ListNotifications retrieves all notifications from the database
func ListNotifications(c *fiber.Ctx) error {
	// Get a handle to the notifications collection
	collection := mymongo.GetMongoClient().Database("seng468_a2_db").Collection("notifications")

	// Find all notifications in the database
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve notifications from database",
		})
	}

	// Decode the cursor into a slice of notifications
	var notifications []models.Notification
	if err := cursor.All(context.Background(), &notifications); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not decode notifications from cursor",
		})
	}

	// Return the notifications
	return c.JSON(notifications)
}

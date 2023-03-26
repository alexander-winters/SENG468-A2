package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

// CreateNotification inserts a new notification into the database
func CreateNotification(c *fiber.Ctx) error {
	// Get a handle to the notifications collection
	notificationCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("notifications")
	userCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Parse the request body into a struct
	var notification models.Notification
	if err := c.BodyParser(&notification); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request body",
		})
	}

	// Set the created time
	notification.CreatedAt = time.Now()

	ctx := context.Background()

	// Create channels for error handling and synchronization
	errChan := make(chan error)
	insertedIDChan := make(chan primitive.ObjectID)

	// Concurrently insert the notification into the database
	go func() {
		res, err := notificationCollection.InsertOne(ctx, notification)
		if err != nil {
			errChan <- err
			return
		}
		insertedIDChan <- res.InsertedID.(primitive.ObjectID)
	}()

	// Concurrently find the user and update their notifications
	go func() {
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"username": notification.Recipient}).Decode(&user)
		if err != nil {
			errChan <- err
			return
		}

		// Add the notification to the user's array of notifications
		user.Notifications = append(user.Notifications, notification)

		// Update the user in the database
		filter := bson.M{"username": notification.Recipient}
		update := bson.M{"$set": bson.M{"notifications": user.Notifications}}
		if _, err := userCollection.UpdateOne(ctx, filter, update); err != nil {
			errChan <- err
			return
		}

		errChan <- nil
	}()

	// Wait for the concurrent tasks to finish and handle any errors
	var insertedID primitive.ObjectID
	select {
	case err := <-errChan:
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not process request",
			})
		}
	case insertedID = <-insertedIDChan:
	}

	// Set the ID of the notification
	notification.ID = insertedID

	// Return the notification
	return c.JSON(notification)
}

// MarkNotificationAsRead updates the read status of a notification in the database by ID
func MarkNotificationAsRead(c *fiber.Ctx) error {
	// Get a handle to the notifications and users collections
	notificationsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("notifications")
	usersCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Get the ID and username from the URL params
	notificationID := c.Params("ID")
	username := c.Params("username")

	// Convert the notification ID to a MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid notification ID",
		})
	}

	// Create a context
	ctx := context.Background()

	// Create channels for error handling and synchronization
	errChan := make(chan error, 2)

	// Concurrently find the user in the database and update the notification status
	go func() {
		var user models.User
		err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
		if err != nil {
			errChan <- err
			return
		}

		// Update the notification status for the user
		for i, notification := range user.Notifications {
			if notification.ID == objID {
				user.Notifications[i].ReadStatus = true
				user.Notifications[i].UpdatedAt = time.Now()
				break
			}
		}

		// Update the user in the database
		_, err = usersCollection.UpdateOne(ctx, bson.M{"username": username}, bson.M{"$set": bson.M{"notifications": user.Notifications}})
		errChan <- err
	}()

	// Concurrently update the notification in the database
	go func() {
		filter := bson.M{"_id": objID}
		update := bson.M{"$set": bson.M{"read_status": true, "updated_at": time.Now()}}
		_, err := notificationsCollection.UpdateOne(ctx, filter, update)
		errChan <- err
	}()

	// Wait for the concurrent tasks to finish and handle any errors
	for i := 0; i < 2; i++ {
		err := <-errChan
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not process request",
			})
		}
	}

	// Return a success message
	return c.JSON(fiber.Map{
		"message": "Notification read status updated successfully",
	})
}

// ListNotifications retrieves all notifications from the database by user
func ListNotifications(c *fiber.Ctx) error {
	// Get a handle to the notifications collection
	notificationCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("notifications")

	// Get the username from the request parameters
	username := c.Params("username")

	// Get the query parameter for read status
	readStatus := c.Query("read_status")

	// Create a filter based on the username and read status
	filter := bson.M{"username": username}
	if readStatus != "" {
		if readStatus == "true" {
			filter["read_status"] = true
		} else if readStatus == "false" {
			filter["read_status"] = false
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid read status",
			})
		}
	}

	// Find all notifications in the database that match the filter
	cursor, err := notificationCollection.Find(context.Background(), filter)
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

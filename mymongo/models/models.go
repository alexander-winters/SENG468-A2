package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the database
type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Username      string             `bson:"username"`
	FirstName     string             `bson:"first_name"`
	LastName      string             `bson:"last_name"`
	Email         string             `bson:"email"`
	Password      string             `bson:"password"`
	DateOfBirth   time.Time          `bson:"date_of_birth"`
	ListOfFriends []string           `bson:"list_of_friends"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}

// Post represents a post in the database
type Post struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	Content          string             `bson:"content" json:"content"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
	NumberOfLikes    int                `bson:"number_of_likes" json:"number_of_likes"`
	NumberOfComments int                `bson:"number_of_comments" json:"number_of_comments"`
}

// Comment represents a comment in the database
type Comment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID        primitive.ObjectID `bson:"post_id" json:"post_id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Content       string             `bson:"content" json:"content"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
	NumberOfLikes int                `bson:"number_of_likes" json:"number_of_likes"`
}

type NotificationType string

const (
	PostCreatedNotification    NotificationType = "post_created"
	CommentCreatedNotification NotificationType = "comment_created"
	PostLikedNotification      NotificationType = "post_liked"
)

// Notification represents a notification in the database
type Notification struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type       NotificationType   `bson:"type" json:"type"`
	PostID     primitive.ObjectID `bson:"post_id" json:"post_id"`
	CommentID  primitive.ObjectID `bson:"comment_id" json:"comment_id"`
	Content    string             `bson:"content" json:"content"`
	ReadStatus bool               `bson:"read_status" json:"read_status"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at,omitempty"`
}

// PostReport represents a report of the number of posts created by each user
type PostReport struct {
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
	Count  int                `bson:"count" json:"count"`
}

// CommentReports represents a report on comments created by a user
type CommentReports struct {
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	TotalComments int                `bson:"total_comments" json:"total_comments"`
}

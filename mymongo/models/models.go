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
	PostCount     int                `bson:"post_count" json:"post_count"`
	Notifications []Notification     `bson:"notifications" json:"notifications"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}

// Post represents a post in the database
type Post struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username         string             `bson:"username" json:"username"`
	PostNumber       int                `bson:"post_number" json:"post_number"`
	Content          string             `bson:"content" json:"content"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
	Comments         []Comment          `bson:"comments" json:"comments"`
	NumberOfLikes    int                `bson:"number_of_likes" json:"number_of_likes"`
	NumberOfComments int                `bson:"number_of_comments" json:"number_of_comments"`
}

// Comment represents a comment in the database
type Comment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID        primitive.ObjectID `bson:"post_id" json:"post_id"`
	PostNumber    int                `bson:"post_number" json:"post_number"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username      string             `bson:"username" json:"username"`
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
	CommentLikedNotification   NotificationType = "comment_liked"
)

// Notification represents a notification in the database
type Notification struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username   string             `bson:"username" json:"username"`
	Type       NotificationType   `bson:"type" json:"type"`
	PostID     primitive.ObjectID `bson:"post_id" json:"post_id"`
	CommentID  primitive.ObjectID `bson:"comment_id" json:"comment_id"`
	Recipient  string             `bson:"recipient" json:"recipient"`
	Content    string             `bson:"content" json:"content"`
	ReadStatus bool               `bson:"read_status" json:"read_status"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}

// PostReport represents a report of the number of posts created by each user
type PostReport struct {
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username string             `bson:"username" json:"username"`
	Count    int                `bson:"count" json:"count"`
}

// CommentReport represents a report on comments created by a user
type CommentReport struct {
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username      string             `bson:"username" json:"username"`
	TotalComments int                `bson:"total_comments" json:"total_comments"`
}

// LikeReports represents a report on likes given or received by a user
type LikeReport struct {
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username      string             `bson:"username" json:"username"`
	LikesGiven    int                `bson:"likes_given" json:"likes_given"`
	LikesReceived int                `bson:"likes_received" json:"likes_received"`
}

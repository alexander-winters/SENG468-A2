package postScripts

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

func CreatePostsForUsers(numPosts int) {
	usernames, err := readUsernamesFromFile("users.txt")
	if err != nil {
		log.Fatalf("Error reading usernames from file: %v", err)
		return
	}

	for _, username := range usernames {
		for i := 0; i < numPosts; i++ {
			createRandomPost(username)
			time.Sleep(50 * time.Millisecond) // Add a short delay to avoid overwhelming the server
		}
	}
}

func readUsernamesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var usernames []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			username := strings.TrimSpace(parts[2])
			usernames = append(usernames, username)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return usernames, nil
}

func createRandomPost(username string) {
	post := &models.Post{
		Username:  username,
		Content:   randomString(10),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	postJSON, err := json.Marshal(post)
	if err != nil {
		fmt.Println("Error marshaling post:", err)
		return
	}

	url := fmt.Sprintf("http://localhost:3000/user/%s/post", username)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(postJSON))
	if err != nil {
		fmt.Println("Error creating post:", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error creating post: status code %d\n", resp.StatusCode)
	} else {
		fmt.Printf("Post created successfully for user %s\n", username)
	}
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

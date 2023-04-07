package commentScripts

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

func CreateCommentsForUsers(numComments int) {
	userPostInfos, err := readUserPostInfoFromFile("posts.txt")
	if err != nil {
		fmt.Printf("Error reading user post info from file: %v", err)
		return
	}

	for _, userPostInfo := range userPostInfos {
		for i := 0; i < numComments; i++ {
			createRandomComment(userPostInfo.Username, userPostInfo.PostNumber)
			time.Sleep(50 * time.Millisecond) // Add a short delay to avoid overwhelming the server
		}
	}
}

type UserPostInfo struct {
	Username   string
	PostNumber int
}

func readUserPostInfoFromFile(filename string) ([]UserPostInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var userPostInfos []UserPostInfo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			username := strings.TrimSpace(parts[1])
			postNumber, err := strconv.Atoi(strings.TrimSpace(parts[2]))
			if err == nil {
				userPostInfos = append(userPostInfos, UserPostInfo{
					Username:   username,
					PostNumber: postNumber,
				})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return userPostInfos, nil
}

func createRandomComment(username string, postNumber int) {
	comment := &models.Comment{
		Username:  username,
		Content:   randomString(10),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	commentJSON, err := json.Marshal(comment)
	if err != nil {
		fmt.Println("Error marshaling comment:", err)
		return
	}

	url := fmt.Sprintf("http://localhost:80/user/%s/post/%d/comment", username, postNumber)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(commentJSON))
	if err != nil {
		fmt.Println("Error creating comment:", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error creating comment: status code %d\n", resp.StatusCode)
	} else {
		fmt.Printf("Comment created successfully for user %s post %d\n", username, postNumber)
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

package likeScripts

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Content struct {
	Username   string
	PostNumber int
	CommentID  string
}

func readFromFile(filename string) ([]Content, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var contents []Content
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")

		if len(parts) < 3 {
			return nil, errors.New("invalid line format")
		}

		postNumber, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			postNumber = 0
		}

		content := Content{
			Username:   strings.TrimSpace(parts[0]),
			PostNumber: postNumber,
			CommentID:  strings.TrimSpace(parts[2]),
		}
		contents = append(contents, content)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return contents, nil
}

func readPostsAndComments(postsFilename, commentsFilename string) ([]Content, error) {
	posts, err := readFromFile(postsFilename)
	if err != nil {
		return nil, err
	}

	comments, err := readFromFile(commentsFilename)
	if err != nil {
		return nil, err
	}

	contents := append(posts, comments...)
	return contents, nil
}

func sendLikeRequest(url string) error {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader("{}"))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to like, status code: %d", resp.StatusCode)
	}

	return nil
}

func randomlyLike(contents []Content) {
	src := rand.NewSource(time.Now().UnixNano())
	rand.New(src)
	likedPosts := make(map[int]bool)
	likedComments := make(map[string]bool)

	for _, content := range contents {
		likePost := rand.Intn(2) == 0
		likeComment := rand.Intn(2) == 0

		if likePost && !likedPosts[content.PostNumber] {
			likedPosts[content.PostNumber] = true
			postURL := fmt.Sprintf("http://localhost:80/user/%s/post/%d/like", content.Username, content.PostNumber)
			err := sendLikeRequest(postURL)
			if err != nil {
				fmt.Printf("Error liking post: %v\n", err)
			} else {
				fmt.Printf("User '%s' likes post number %d\n", content.Username, content.PostNumber)
			}
		}

		if likeComment && !likedComments[content.CommentID] {
			likedComments[content.CommentID] = true
			commentURL := fmt.Sprintf("http://localhost:80/user/%s/post/%d/comment/%s/like", content.Username, content.PostNumber, content.CommentID)
			err := sendLikeRequest(commentURL)
			if err != nil {
				fmt.Printf("Error liking comment: %v\n", err)
			} else {
				fmt.Printf("User '%s' likes comment with ID %s\n", content.Username, content.CommentID)
			}
		}

	}
}

func LikePostsAndComments(postsFilename, commentsFilename string) {
	contents, err := readPostsAndComments(postsFilename, commentsFilename)
	if err != nil {
		fmt.Printf("Error reading contents: %v\n", err)
		return
	}
	randomlyLike(contents)
}

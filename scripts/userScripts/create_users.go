package userScripts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/alexander-winters/SENG468-A2/mymongo/models"
)

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func randomDate(minYear int, maxYear int) time.Time {
	year := rand.Intn(maxYear-minYear+1) + minYear
	month := rand.Intn(12) + 1
	day := rand.Intn(28) + 1 // Assume up to 28 days per month to simplify
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func generateRandomUser() models.User {
	user := models.User{
		Username:    randomString(10),
		FirstName:   randomString(5),
		LastName:    randomString(7),
		Email:       fmt.Sprintf("%s@example.com", randomString(10)),
		Password:    randomString(12),
		DateOfBirth: randomDate(1940, 2010),
	}
	return user
}

func CreateRandomUsers(createUsers int) {
	for i := 0; i < createUsers; i++ {
		user := generateRandomUser()

		userJSON, err := json.Marshal(user)
		if err != nil {
			fmt.Println("Error marshaling user:", err)
			continue
		}

		resp, err := http.Post("http://localhost:80/user", "application/json", bytes.NewBuffer(userJSON))
		if err != nil {
			fmt.Println("Error creating user:", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error creating user: status code %d\n", resp.StatusCode)
		} else {
			fmt.Printf("User %d created successfully\n", i+1)
		}
		time.Sleep(50 * time.Millisecond) // Add a short delay to avoid overwhelming the server
	}
}

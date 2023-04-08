package reportScripts

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadReports(username string) {
	urls := []string{
		fmt.Sprintf("http://localhost:80/reports/%s/posts", username),
		fmt.Sprintf("http://localhost:80/reports/%s/comments", username),
		fmt.Sprintf("http://localhost:80/reports/%s/likes", username),
	}

	filenames := []string{
		username + "_posts.json",
		username + "_comments.json",
		username + "_likes.json",
	}

	// Create the reports sub-directory if it doesn't exist
	reportsDir := "reports"
	if _, err := os.Stat(reportsDir); os.IsNotExist(err) {
		if err := os.Mkdir(reportsDir, 0755); err != nil {
			log.Fatalf("Error creating reports directory: %v", err)
		}
	}

	for i, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching data from %s: %v", url, err)
			continue
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading data from %s: %v", url, err)
			continue
		}

		var jsonData interface{}
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			log.Printf("Error unmarshalling JSON data from %s: %v", url, err)
			continue
		}

		prettyJson, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			log.Printf("Error formatting JSON data from %s: %v", url, err)
			continue
		}

		// Add the sub-directory path to the filenames
		filepathInReportsDir := filepath.Join(reportsDir, filenames[i])

		err = os.WriteFile(filepathInReportsDir, prettyJson, 0644)
		if err != nil {
			log.Printf("Error writing JSON data to %s: %v", filepathInReportsDir, err)
			continue
		}

		fmt.Printf("Data from %s saved to %s\n", url, filepathInReportsDir)
	}
}

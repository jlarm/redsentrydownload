package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"redsentry.joelohr.com/auth"
)

func main() {
	token, err := auth.GetValidToken()
	if err != nil {
		log.Fatalf("Failed to get valid token: %v", err)
	}

	fmt.Print("Enter Sentry ID: ")
	reader := bufio.NewReader(os.Stdin)
	id, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	id = strings.TrimSpace(id)

	if id == "" {
		log.Fatalf("ID cannot be empty")
	}

	baseURL := os.Getenv("REDSENTRY_API_URL")
	if baseURL == "" {
		log.Fatalf("REDSENTRY_API_URL environment variable is not set")
	}

	url := fmt.Sprintf("%s/scanners/external/%s/scan/dates", baseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to get scan dates. Status: %d\nBody: %s\n", resp.StatusCode, string(body))
	}

	type ScanResult struct {
		ID     int    `json:"id"`
		Date   string `json:"date"`
		Status string `json:"status"`
	}

	var scanResults []ScanResult
	if err := json.Unmarshal(body, &scanResults); err != nil {
		log.Fatalf("Failed to parse JSON response %v", err)
	}

	var scanIDs []int
	for _, result := range scanResults {
		if result.Status == "done" {
			scanIDs = append(scanIDs, result.ID)
		}
	}

	fmt.Println("Scan IDs:", scanIDs)
}

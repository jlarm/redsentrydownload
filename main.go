package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	fmt.Print("Select scanner type (internal/external): ")
	scannerType, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	scannerType = strings.TrimSpace(scannerType)

	if scannerType != "internal" && scannerType != "external" {
		log.Fatalf("Invalid scanner type. Must be 'internal' or 'external'")
	}

	fmt.Print("Select report type (executive/technical): ")
	reportType, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	reportType = strings.TrimSpace(reportType)

	if reportType != "executive" && reportType != "technical" {
		log.Fatalf("Invalid report type. Must be 'executive' or 'technical'")
	}

	baseURL := os.Getenv("REDSENTRY_API_URL")
	if baseURL == "" {
		log.Fatalf("REDSENTRY_API_URL environment variable is not set")
	}

	url := fmt.Sprintf("%s/scanners/%s/%s/scan/dates", baseURL, scannerType, id)

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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}

	reportDir := filepath.Join(homeDir, "Downloads", fmt.Sprintf("reports_%s", id))
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		log.Fatalf("Failed to create report directory: %v", err)
	}

	fmt.Printf("Downloading PDF reports to folder: %v", reportDir)
	for _, scanID := range scanIDs {
		downloadPDFReport(token, id, scanID, reportDir, reportType)
	}

	fmt.Println("All reports downloaded successfully")
}

func downloadPDFReport(token, sentryID string, scanID int, outputDir string, reportType string) {
	baseURL := os.Getenv("REDSENTRY_API_URL")
	url := fmt.Sprintf("%s/scanners/external/%s/report/%s?format=pdf&scan_id=%d", baseURL, sentryID, reportType, scanID)

	fmt.Printf("Downloading report for scan ID %d...\n", scanID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create request for scan ID %d: %v", scanID, err)
		return
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request for scan ID %d: %v", scanID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to download report for scan ID %d. Status: %d, Body: %s", scanID, resp.StatusCode, string(body))
		return
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("report_%s_scan_%d.pdf", sentryID, scanID))
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Failed to create output file for scan ID %d: %v", scanID, err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		log.Printf("Failed to write PDF data for scan ID %d: %v", scanID, err)
		return
	}

	fmt.Printf("PDF report downloaded successfully for scan ID %d\n", scanID)
}

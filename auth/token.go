package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Credentials struct {
	Username string
	Password string
}

type TokenInfo struct {
	Token     string
	ExpiresAt time.Time
}

var (
	tokenCache TokenInfo
	tokenMutex sync.RWMutex
)

func LoadCredentials() (Credentials, error) {
	_ = godotenv.Load()

	username := os.Getenv("REDSENTRY_USERNAME")
	password := os.Getenv("REDSENTRY_PASSWORD")

	if username == "" || password == "" {
		return Credentials{}, fmt.Errorf("missing required REDSENTRY_USERNAME or REDSENTRY_PASSWORD environment variables")
	}

	return Credentials{
		Username: username,
		Password: password,
	}, nil
}

func GetToken(creds Credentials) (string, error) {
	url := os.Getenv("REDSENTRY_LOGIN_URL")
	if url == "" {
		return "", fmt.Errorf("missing required REDSENTRY_LOGIN_URL environment variable")
	}

	credentials := map[string]string{
		"username": creds.Username,
		"password": creds.Password,
	}

	jsonData, err := json.Marshal(credentials)
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed. Status: %d\nBody: %s\n", resp.StatusCode, string(body))
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(body, &responseBody); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %v", err)
	}

	token, ok := responseBody["token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	expiresIn := 3600.0
	if exp, ok := responseBody["expires_in"].(float64); ok {
		expiresIn = exp
	}

	tokenMutex.Lock()
	tokenCache = TokenInfo{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	tokenMutex.Unlock()

	return token, nil
}

func IsTokenValid() bool {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()

	if tokenCache.Token == "" || time.Now().After(tokenCache.ExpiresAt) {
		return false
	}

	bufferTime := 30 * time.Second
	return time.Now().Add(bufferTime).Before(tokenCache.ExpiresAt)
}

func GetValidToken() (string, error) {
	if IsTokenValid() {
		tokenMutex.RLock()
		token := tokenCache.Token
		tokenMutex.RUnlock()
		return token, nil
	}

	creds, err := LoadCredentials()
	if err != nil {
		return "", fmt.Errorf("failed to load credentials: %v", err)
	}

	return GetToken(creds)
}

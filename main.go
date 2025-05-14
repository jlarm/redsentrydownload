package main

import (
	"fmt"
	"log"

	"redsentry.joelohr.com/auth"
)

func main() {
	token, err := auth.GetValidToken()
	if err != nil {
		log.Fatalf("Failed to get valid token: %v", err)
	}

	fmt.Println("Access Token: ", token)
}

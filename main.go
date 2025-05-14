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

	if auth.IsTokenValid() {
		fmt.Println("Token is valid")
	} else {
		fmt.Println("Token is not valid")
	}
}

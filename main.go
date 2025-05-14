package main

import (
	"fmt"
	"log"

	"redsentry.joelohr.com/auth"
)

func main() {
	creds, err := auth.LoadCredentials()
	if err != nil {
		log.Fatalf("Failed to load credentials: %v", err)
	}

	token, err := auth.GetToken(creds)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Access Token: ", token)
}

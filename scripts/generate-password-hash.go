package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate-password-hash.go <password>")
		os.Exit(1)
	}
	
	password := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", string(hash))
}


package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate-password.go <password>")
		fmt.Println("Example: go run generate-password.go admin123")
		return
	}

	password := os.Args[1]

	// Generate hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}

	fmt.Println("========================================")
	fmt.Println("Password Hash Generator")
	fmt.Println("========================================")
	fmt.Println("Password:", password)
	fmt.Println("Hash:", string(hash))
	fmt.Println("")
	fmt.Println("SQL Update Command:")
	fmt.Printf("UPDATE users SET password = '%s' WHERE username = 'admin';\n", string(hash))
	fmt.Println("========================================")

	// Verify hash
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err == nil {
		fmt.Println("✅ Hash verification: SUCCESS")
	} else {
		fmt.Println("❌ Hash verification: FAILED")
	}
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Connect to DB
	dsn := "host=localhost port=5432 user=mazadpay password=mazadpay_secret dbname=mazadpay sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		fmt.Println("DB connection error:", err)
		os.Exit(1)
	}

	// Generate new hash
	pin := "0000"
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Hash error:", err)
		os.Exit(1)
	}

	fmt.Println("New hash:", string(hash))

	// Update DB
	_, err = db.ExecContext(context.Background(), "UPDATE users SET password_hash = $1 WHERE phone = $2", string(hash), "22222222")
	if err != nil {
		fmt.Println("Update error:", err)
		os.Exit(1)
	}
	fmt.Println("Updated successfully!")

	// Verify
	var storedHash string
	err = db.Get(&storedHash, "SELECT password_hash FROM users WHERE phone = '22222222'")
	if err != nil {
		fmt.Println("Verify error:", err)
		os.Exit(1)
	}
	fmt.Println("Stored hash:", storedHash)

	// Test match
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("0000"))
	fmt.Println("Match:", err == nil)
}
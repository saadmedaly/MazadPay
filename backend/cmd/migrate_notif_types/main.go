package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "mazadpay"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "mazadpay_secret"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "mazadpay"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Drop old constraint
	_, err = db.Exec("ALTER TABLE notifications DROP CONSTRAINT IF EXISTS chk_notif_type")
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	// Add new constraint
	_, err = db.Exec("ALTER TABLE notifications ADD CONSTRAINT chk_notif_type CHECK (type IN ('bid', 'win', 'payment', 'system', 'ad', 'general', 'new_auction', 'transaction', 'report'))")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Migration applied successfully!")
}

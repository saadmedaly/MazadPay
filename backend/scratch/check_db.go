package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func scractch() {
	godotenv.Load("../.env")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	ssl := os.Getenv("DB_SSL_MODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, ssl)
	
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	var encoding string
	err = db.Get(&encoding, "SHOW server_encoding")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server Encoding: %s\n", encoding)

	err = db.Get(&encoding, "SHOW client_encoding")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Client Encoding: %s\n", encoding)

    // Check database character set
    var dbname, dbencoding, dbcollate string
    err = db.QueryRow("SELECT datname, pg_encoding_to_char(encoding), datcollate FROM pg_database WHERE datname = $1", name).Scan(&dbname, &dbencoding, &dbcollate)
    if err != nil {
        fmt.Printf("Could not check pg_database: %v\n", err)
    } else {
        fmt.Printf("Database %s: Encoding=%s Collation=%s\n", dbname, dbencoding, dbcollate)
    }

	// Check a sample category
	var nameAr string
	err = db.Get(&nameAr, "SELECT name_ar FROM categories LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sample name_ar: %s\n", nameAr)
}

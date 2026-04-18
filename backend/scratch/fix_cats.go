package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	ssl := os.Getenv("DB_SSL_MODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s client_encoding=UTF8",
		host, port, user, pass, name, ssl)
	
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Update existing NULLs to empty strings
	_, err = db.Exec("UPDATE categories SET name_en = '' WHERE name_en IS NULL")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Make column NOT NULL and set DEFAULT
	_, err = db.Exec("ALTER TABLE categories ALTER COLUMN name_en SET NOT NULL")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("ALTER TABLE categories ALTER COLUMN name_en SET DEFAULT ''")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database fix completed: name_en is now NOT NULL")
}

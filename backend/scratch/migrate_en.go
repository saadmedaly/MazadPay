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

	_, err = db.Exec("ALTER TABLE categories ADD COLUMN IF NOT EXISTS name_en VARCHAR(100)")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Migration completed: name_en added to categories")
}

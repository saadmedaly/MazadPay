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

	queries := []string{
		"ALTER TABLE locations RENAME COLUMN city_name TO city_name_ar",
		"ALTER TABLE locations RENAME COLUMN area_name TO area_name_ar",
		"ALTER TABLE locations ADD COLUMN city_name_fr VARCHAR(100)",
		"ALTER TABLE locations ADD COLUMN area_name_fr VARCHAR(100)",
		"UPDATE locations SET city_name_fr = city_name_ar, area_name_fr = area_name_ar",
		"ALTER TABLE locations ALTER COLUMN city_name_fr SET NOT NULL",
		"ALTER TABLE locations ALTER COLUMN city_name_fr SET DEFAULT ''",
		"ALTER TABLE locations ALTER COLUMN area_name_fr SET DEFAULT ''",
	}

	for _, q := range queries {
		fmt.Printf("Executing: %s\n", q)
		_, err = db.Exec(q)
		if err != nil {
			log.Printf("Warning: %v\n", err)
		}
	}

	fmt.Println("Migration completed: locations updated")
}

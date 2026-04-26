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

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Connected to database, applying locations schema migration...")

	// Migration SQL
	migrationSQL := `
-- Vérifier et renommer si les colonnes existent
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'locations' AND column_name = 'city_name') THEN
        ALTER TABLE locations RENAME COLUMN city_name TO city_name_ar;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'locations' AND column_name = 'area_name') THEN
        ALTER TABLE locations RENAME COLUMN area_name TO area_name_ar;
    END IF;
END $$;

-- Ajouter les colonnes FR si elles n'existent pas
ALTER TABLE locations ADD COLUMN IF NOT EXISTS city_name_fr VARCHAR(100);
ALTER TABLE locations ADD COLUMN IF NOT EXISTS area_name_fr VARCHAR(100);

-- Initialiser les valeurs FR avec les valeurs AR pour éviter les NULL
UPDATE locations SET city_name_fr = city_name_ar, area_name_fr = area_name_ar WHERE city_name_fr IS NULL;

ALTER TABLE locations ALTER COLUMN city_name_fr SET NOT NULL;
ALTER TABLE locations ALTER COLUMN city_name_fr SET DEFAULT '';
ALTER TABLE locations ALTER COLUMN area_name_fr SET DEFAULT '';
`

	_, err = db.Exec(migrationSQL)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("Locations schema migration applied successfully!")
}

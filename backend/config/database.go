package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func ConnectDatabase() (*sql.DB, error) {

	// Membaca konfigurasi .env
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file .env: %w", err)
	}

	// Mengambil konfigurasi database
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Membuat konfigurasi koneksi
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user,
		password,
		host,
		port,
		dbname,
	)

	// Membuka koneksi database
	db, err := sql.Open("mysql", dsn)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err != nil {
		return nil, err
	}

	// Memeriksa koneksi database
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

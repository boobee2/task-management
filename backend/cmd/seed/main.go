package main

import (
	"database/sql"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"task-management-system/config"
	"task-management-system/internal/model"
	"task-management-system/internal/repository"
)

func main() {
	// ConnectDatabase sudah memuat .env, jadi panggil ini lebih dulu
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("gagal konek database: ", err)
	}
	defer db.Close()

	name := os.Getenv("ADMIN_NAME")
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if name == "" || email == "" || password == "" {
		log.Fatal("ADMIN_NAME, ADMIN_EMAIL, dan ADMIN_PASSWORD wajib diisi di .env")
	}

	repo := repository.NewUserRepository(db)

	// jangan buat dobel kalau email sudah ada
	_, err = repo.GetByEmail(email)
	if err == nil {
		log.Println("admin dengan email tersebut sudah ada, seeder dilewati")
		return
	}
	if err != sql.ErrNoRows {
		log.Fatal("gagal cek email: ", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("gagal hash password: ", err)
	}

	admin := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		Role:     "ADMIN",
		IsActive: true,
	}
	if err := repo.Create(admin); err != nil {
		log.Fatal("gagal membuat admin: ", err)
	}

	log.Println("admin berhasil dibuat:", email)
}

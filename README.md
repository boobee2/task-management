# Task Management System

Aplikasi web manajemen tugas (*full-stack*) yang dibangun menggunakan Go (Golang) untuk REST API di sisi backend serta Vanilla HTML, CSS, dan JavaScript di sisi frontend, lengkap dengan sistem autentikasi dan manajemen hak akses (Admin & User).

## 📂 Struktur Proyek

TaskManagementSystem/
├── backend/          # REST API (Go, Clean Architecture, MySQL)
│   ├── cmd/          # Entrypoint aplikasi & seeder database
│   ├── config/       # Konfigurasi database & lingkungan
│   ├── internal/     # Core logic (DTO, Handler, Middleware, Model, Repository, Service)
│   ├── migrations/   # Skrip SQL migrasi database
│   └── routes/       # Definisi rute endpoint API
└── frontend/         # Antarmuka web (HTML, CSS, JavaScript)
    ├── css/          # Lembar gaya tampilan
    └── js/           # Logika aplikasi & komunikasi API
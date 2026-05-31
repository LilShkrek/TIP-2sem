package main

import (
	"crypto/tls"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"example.com/pz5-security/internal/config"
	"example.com/pz5-security/internal/httpapi"
	"example.com/pz5-security/internal/student"
)

func main() {
	cfg := config.New()

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	repo := student.NewRepo(db)

	stmt, err := repo.PrepareGetByID()
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	handler := httpapi.NewHandler(repo, stmt)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/students", handler.GetStudentByID)

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	log.Printf("HTTPS server started on https://localhost%s", cfg.Addr)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}

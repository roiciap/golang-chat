package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/roiciap/golang-chat/internal/data/db/migrations"
	"github.com/roiciap/golang-chat/internal/http/handlers"
	myauth "github.com/roiciap/golang-chat/pkg/auth"
)

func initHttpHandler() (http.Handler, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		return nil, errors.New("aha")
	}

	authHandler := myauth.CreateJWTAuthentication([]byte(jwtSecret))
	handler := handlers.CreateAccountHandler(authHandler)
	return handler, nil
}

func main() {
	err := migrations.Migrate()
	if err != nil {
		log.Fatal("Problem migrating database")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	handler, err := initHttpHandler()

	if err != nil {
		log.Fatal("No JWT_SECRET provided")
		os.Exit(1)
	}

	mux.Handle("/login", handler)
	mux.Handle("/register", handler)
	mux.Handle("/settings/", handler)

	http.ListenAndServe(":8080", mux)
}

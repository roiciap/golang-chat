package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/roiciap/golang/internal/http/handlers"
	myauth "github.com/roiciap/golang/pkg/auth"
)

func initHttpHandler() (http.Handler, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		return nil, errors.New("aha")
	}

	authHandler := myauth.CreateJWTAuthentication([]byte(jwtSecret))
	handler, err := handlers.CreateAccountHandler(authHandler)
	if err != nil {
		fmt.Println(err.Error())
	}
	return handler, nil
}

func main() {
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

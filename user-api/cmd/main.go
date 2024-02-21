package main

import (
	"fmt"
	"net/http"

	"github.com/roiciap/golang/myauth"
	"github.com/roiciap/golang/user-api/handlers"
)

func main() {
	mux := http.NewServeMux()
	authHandler := myauth.CreateJWTAuthentication([]byte("secret"))
	handler, err := handlers.CreateAccountHandler(authHandler)
	if err != nil {
		fmt.Println(err.Error())
	}

	mux.Handle("/login", handler)
	mux.Handle("/register", handler)

	http.ListenAndServe(":8080", mux)
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/validator.v2"
)

type Credentials struct {
	Login    string `json:"login" validate:"min=1"`
	Password string `json:"password" validate:"min=1"`
}

var CredentialsDatabase = []Credentials{
	{Login: "mateusz", Password: "Bazior"},
}

func handleAuthentication(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Błąd w odczycie danych uwierzytelniających", http.StatusBadRequest)
		return
	}
	if err := validator.Validate(creds); err != nil {
		http.Error(w, "Niepoprawne body zapytania", http.StatusBadRequest)
		return
	}

	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Błąd podczas tworzenia odpowiedzi", http.StatusInternalServerError)
	}
	w.Write(jsonResp)
}

func main() {
	// Ustawienie endpointu dla uwierzytelniania
	http.HandleFunc("/login", handleAuthentication)

	// Uruchomienie serwera na porcie 8080
	fmt.Println("Serwer uruchomiony na porcie 8080")
	http.ListenAndServe(":8080", nil)
}

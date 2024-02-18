package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"

	"gopkg.in/validator.v2"
)

var (
	loginRegex    = regexp.MustCompile(`^\/login$`)
	registerRegex = regexp.MustCompile(`^\/register$`)
)

type Credentials struct {
	Login    string `json:"login" validate:"min=1"`
	Password string `json:"password" validate:"min=1"`
}

type CredDatastore struct {
	m      map[int]Credentials
	nextId int `default:"1"`
	*sync.RWMutex
}

type AccountHandler struct {
	store *CredDatastore
}

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodPost && loginRegex.MatchString(r.URL.Path):
		h.Login(w, r)
		return
	case r.Method == http.MethodPost && registerRegex.MatchString(r.URL.Path):
		h.Register(w, r)
		return
	default:
		notFound(w, r)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error": "not found"}`))
}

func (h *AccountHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := readCredsFromBody(r.Body, &creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.RLock()
	if h.findCredId(creds) == -1 {
		h.store.RUnlock()
		http.Error(w, "User doesn't exist!", http.StatusConflict)
		return
	}
	fmt.Println("dobry login")
	h.store.RUnlock()

	w.WriteHeader(http.StatusOK)
}
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := readCredsFromBody(r.Body, &creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	h.store.Lock()
	if h.findCredId(creds) != -1 {
		h.store.Unlock()
		http.Error(w, "User already exists!", http.StatusConflict)
		return
	}
	h.store.m[h.store.nextId] = creds
	h.store.nextId++
	h.store.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) findCredId(creds Credentials) int {
	for id, value := range h.store.m {
		if value.Login == creds.Login {
			return id
		}
	}
	return -1
}

func readCredsFromBody(body io.ReadCloser, creds *Credentials) error {

	err := json.NewDecoder(body).Decode(creds)
	if err != nil {
		return err
	}
	if err := validator.Validate(*creds); err != nil {
		return err
	}
	return nil
}

func main() {
	mux := http.NewServeMux()

	handler := &AccountHandler{
		store: &CredDatastore{
			m:       map[int]Credentials{},
			RWMutex: &sync.RWMutex{},
		},
	}

	mux.Handle("/login", handler)
	mux.Handle("/register", handler)

	http.ListenAndServe(":8080", mux)
}

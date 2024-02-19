package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sync"

	"golang.org/x/crypto/bcrypt"
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

type Account struct {
	Login        string
	PasswordHash []byte
}

type CredDatastore struct {
	m      map[int]Account
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
	err = h.checkUserCreds(creds)
	h.store.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := readCredsFromBody(r.Body, &creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	h.store.RWMutex.Lock()
	err = h.addUser(creds)
	h.store.RWMutex.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) addUser(creds Credentials) error {
	if h.findCredId(creds) != -1 {
		return errors.New("user already exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 14)
	if err != nil {
		return errors.New("problem crating user")
	}
	newUser := Account{creds.Login, hash}
	h.store.m[h.store.nextId] = newUser
	h.store.nextId++
	return nil
}

func (h *AccountHandler) checkUserCreds(creds Credentials) error {
	accId := h.findCredId(creds)
	if accId == -1 {
		return errors.New("user doesnt exist")
	}

	acc := h.store.m[accId]

	err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(creds.Password))
	if err != nil {
		return errors.New("password doesnt match")
	}
	return nil
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
			m:       map[int]Account{},
			RWMutex: &sync.RWMutex{},
		},
	}

	mux.Handle("/login", handler)
	mux.Handle("/register", handler)

	http.ListenAndServe(":8080", mux)
}

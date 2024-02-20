package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/roiciap/golang/user-api/model"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
)

var (
	loginRegex    = regexp.MustCompile(`^\/login$`)
	registerRegex = regexp.MustCompile(`^\/register$`)
)

type AccountHandler struct {
	Store *credDatastore
}

func CreateAccountHandler(creds ...model.Credentials) (*AccountHandler, error) {
	handler := &AccountHandler{
		Store: &credDatastore{
			Database: map[int]model.Account{},
			RWMutex:  &sync.RWMutex{},
		},
	}
	var errs error
	for _, credsItem := range creds {
		err := handler.addUser(credsItem)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return handler, errs
}

type credDatastore struct {
	Database map[int]model.Account
	nextId   int `default:"1"`
	*sync.RWMutex
}

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodPost && loginRegex.MatchString(r.URL.Path):
		h.login(w, r)
		return
	case r.Method == http.MethodPost && registerRegex.MatchString(r.URL.Path):
		h.register(w, r)
		return
	default:
		notFound(w, r)
	}
}

func (h *AccountHandler) login(w http.ResponseWriter, r *http.Request) {
	var creds model.Credentials
	err := readCredsFromBody(r.Body, &creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.Store.RLock()
	err = h.checkUserCreds(creds)
	h.Store.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) register(w http.ResponseWriter, r *http.Request) {
	var creds model.Credentials
	err := readCredsFromBody(r.Body, &creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.Store.RWMutex.Lock()
	err = h.addUser(creds)
	h.Store.RWMutex.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) addUser(creds model.Credentials) error {
	if h.findCredId(creds) != -1 {
		return errors.New("user  " + creds.Login + " already exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 14)
	if err != nil {
		return errors.New("problem creating user")
	}
	newUser := model.Account{Login: creds.Login, PasswordHash: hash}
	h.Store.Database[h.Store.nextId] = newUser
	h.Store.nextId++
	return nil
}

func (h *AccountHandler) checkUserCreds(creds model.Credentials) error {
	accId := h.findCredId(creds)
	if accId == -1 {
		return errors.New("user doesnt exist")
	}

	acc := h.Store.Database[accId]

	err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(creds.Password))
	if err != nil {
		return errors.New("password doesnt match")
	}
	return nil
}

func (h *AccountHandler) findCredId(creds model.Credentials) int {
	for id, value := range h.Store.Database {
		if value.Login == creds.Login {
			return id
		}
	}
	return -1
}

func readCredsFromBody(body io.ReadCloser, creds *model.Credentials) error {

	err := json.NewDecoder(body).Decode(creds)
	if err != nil {
		return err
	}
	if err := validator.Validate(*creds); err != nil {
		return err
	}
	return nil
}

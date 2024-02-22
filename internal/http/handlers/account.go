package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sync"

	domain "github.com/roiciap/golang/internal/business/domains"
	"github.com/roiciap/golang/internal/http/datatransfers/requests"
	myauth "github.com/roiciap/golang/pkg/auth"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
)

var (
	loginRegex    = regexp.MustCompile(`^\/login$`)
	registerRegex = regexp.MustCompile(`^\/register$`)
)

type AccountHandler struct {
	Store        *credDatastore
	AuthStrategy *myauth.IAuthenticationStrategy
}

func CreateAccountHandler(authStrategy myauth.IAuthenticationStrategy, creds ...requests.UserRequest) (*AccountHandler, error) {
	handler := &AccountHandler{
		Store: &credDatastore{
			Database: map[int]domain.UserDomain{},
			RWMutex:  &sync.RWMutex{},
			nextId:   1,
		},
		AuthStrategy: &authStrategy,
	}

	var errs error
	for _, credsItem := range creds {
		_, err := handler.addUser(credsItem)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return handler, errs
}

type credDatastore struct {
	Database map[int]domain.UserDomain
	nextId   int
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
	var creds requests.UserRequest
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
	var creds requests.UserRequest
	err := readCredsFromBody(r.Body, &creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.Store.RWMutex.Lock()
	_, err = h.addUser(creds)
	h.Store.RWMutex.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) addUser(creds requests.UserRequest) (int, error) {
	if h.findCredId(creds) != -1 {
		return -1, errors.New("user  " + creds.Login + " already exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 14)
	if err != nil {
		return -1, errors.New("problem creating user")
	}
	newUser := domain.UserDomain{Login: creds.Login, PasswordHash: hash}
	userId := h.Store.nextId
	h.Store.Database[userId] = newUser
	h.Store.nextId++
	return userId, nil
}

func (h *AccountHandler) checkUserCreds(creds requests.UserRequest) error {
	accId := h.findCredId(creds)
	if accId == -1 {
		return errors.New("user doesnt exist")
	}

	acc := h.Store.Database[accId]

	err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(creds.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return errors.New("passwords doesnt match")
		}
		return errors.New("internal: problem in comparing passwords")
	}
	return nil
}

func (h *AccountHandler) findCredId(creds requests.UserRequest) int {
	for id, value := range h.Store.Database {
		if value.Login == creds.Login {
			return id
		}
	}
	return -1
}

func readCredsFromBody(body io.ReadCloser, creds *requests.UserRequest) error {

	err := json.NewDecoder(body).Decode(creds)
	if err != nil {
		return err
	}
	if err := validator.Validate(*creds); err != nil {
		return err
	}
	return nil
}

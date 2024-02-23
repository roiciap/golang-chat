package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"

	domain "github.com/roiciap/golang-chat/internal/business/domains"
	"github.com/roiciap/golang-chat/internal/http/datatransfers/requests"
	myauth "github.com/roiciap/golang-chat/pkg/auth"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
)

var (
	loginRegex           = regexp.MustCompile(`^\/login$`)
	registerRegex        = regexp.MustCompile(`^\/register$`)
	getUserSettingsRegex = regexp.MustCompile(`^\/settings\/(\d+)$`)
)

type AccountHandler struct {
	Store        *credDatastore
	AuthStrategy myauth.IAuthenticationStrategy
}

func CreateAccountHandler(authStrategy myauth.IAuthenticationStrategy, creds ...requests.AccountRequest) (*AccountHandler, error) {
	handler := &AccountHandler{
		Store: &credDatastore{
			Database: map[int]domain.AccountDomain{},
			RWMutex:  &sync.RWMutex{},
			nextId:   1,
		},
		AuthStrategy: authStrategy,
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
	Database map[int]domain.AccountDomain
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
	case r.Method == http.MethodGet && getUserSettingsRegex.MatchString(r.URL.Path):
		h.getSettings(w, r)
		return
	default:
		notFound(w, r)
	}
}

func (h *AccountHandler) login(w http.ResponseWriter, r *http.Request) {
	var creds requests.AccountRequest
	err := readCredsFromBody(r.Body, &creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.Store.RLock()
	id, err := h.checkUserCreds(creds)
	h.Store.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.AuthStrategy.LoginUser(w, myauth.AuthenticationContext{
		UserId: id,
	})

	if err != nil {
		http.Error(w, "problem signing in, try again later", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) register(w http.ResponseWriter, r *http.Request) {
	var creds requests.AccountRequest
	err := readCredsFromBody(r.Body, &creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.Store.RWMutex.Lock()
	newUserId, err := h.addUser(creds)
	h.Store.RWMutex.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.AuthStrategy.LoginUser(w, myauth.AuthenticationContext{
		UserId: newUserId,
	})

	if err != nil {
		http.Error(w, "Account got created, try to sign in", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) getSettings(w http.ResponseWriter, r *http.Request) {
	matches := getUserSettingsRegex.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}

	authContext, err := h.AuthStrategy.Authenticate(w, r)
	if err != nil {
		return
	}

	userId, err := strconv.Atoi(matches[1])
	if err != nil {
		http.Error(w, "Bad user ID", http.StatusBadRequest)
		return
	}

	h.Store.RLock()
	user, ok := h.Store.Database[userId]
	h.Store.RUnlock()

	if !ok {
		http.Error(w, "Couldnt find user", http.StatusBadRequest)
		return
	}

	if userId != authContext.UserId {
		http.Error(w, "You cant see other users settings !", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"response": "This is private data of user ` + user.Login + `"}`))
}

///

func (h *AccountHandler) addUser(creds requests.AccountRequest) (int, error) {
	if h.findCredId(creds) != -1 {
		return -1, errors.New("user  " + creds.Login + " already exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 14)
	if err != nil {
		return -1, errors.New("problem creating user")
	}
	newUser := domain.AccountDomain{Login: creds.Login, PasswordHash: hash}
	userId := h.Store.nextId
	h.Store.Database[userId] = newUser
	h.Store.nextId++
	return userId, nil
}

func (h *AccountHandler) checkUserCreds(creds requests.AccountRequest) (int, error) {
	accId := h.findCredId(creds)
	if accId == -1 {
		return accId, errors.New("user doesnt exist")
	}

	acc := h.Store.Database[accId]

	err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(creds.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return accId, errors.New("passwords doesnt match")
		}
		return accId, errors.New("internal: problem in comparing passwords")
	}
	return accId, nil
}

func (h *AccountHandler) findCredId(creds requests.AccountRequest) int {
	for id, value := range h.Store.Database {
		if value.Login == creds.Login {
			return id
		}
	}
	return -1
}

func readCredsFromBody(body io.ReadCloser, creds *requests.AccountRequest) error {
	err := json.NewDecoder(body).Decode(creds)
	if err != nil {
		return err
	}
	if err := validator.Validate(*creds); err != nil {
		return err
	}
	return nil
}

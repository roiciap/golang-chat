package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/roiciap/golang-chat/internal/data/crud"
	dbmodels "github.com/roiciap/golang-chat/internal/data/models"
	requests "github.com/roiciap/golang-chat/internal/http/datatransfers/dto"
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
	AuthStrategy myauth.IAuthenticationStrategy
}

func CreateAccountHandler(authStrategy myauth.IAuthenticationStrategy) *AccountHandler {
	handler := &AccountHandler{
		AuthStrategy: authStrategy,
	}

	return handler
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

	id, err := h.getUserId(creds)

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

	accDto, err := bcryptUserCreds(creds)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newUserId, err := crud.AddAccount(accDto)

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

	userId, err := readUserIdFromUrl(matches[1])
	if err != nil {
		http.Error(w, "Bad user ID", http.StatusBadRequest)
		return
	}

	user, err := crud.GetAccountById(userId)

	if err != nil {
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

func (h *AccountHandler) getUserId(creds requests.AccountRequest) (uint, error) {
	acc, err := crud.GetAccountByLogin(creds.Login)
	if err != nil {
		return 0, errors.New("user doesnt exist")
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(creds.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return 0, errors.New("passwords doesnt match")
		}
		return 0, errors.New("internal: problem in comparing passwords")
	}
	return acc.ID, nil
}

func bcryptUserCreds(acc requests.AccountRequest) (dbmodels.AccountDbCreate, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(acc.Password), 14)
	if err != nil {
		return dbmodels.AccountDbCreate{}, err
	}

	accDomain := dbmodels.AccountDbCreate{
		Login:        acc.Login,
		PasswordHash: passwordHash,
	}

	return accDomain, nil
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

func readUserIdFromUrl(urlPart string) (uint, error) {
	userId, err := strconv.Atoi(urlPart)
	if err != nil {
		return 0, err
	}
	if userId < 0 {
		return 0, err
	}
	return uint(userId), nil
}

package myauth

import "net/http"

type IAuthenticationStrategy interface {
	LoginUser(w http.ResponseWriter, context AuthenticationContext) error
	Authenticate(w http.ResponseWriter, r *http.Request) (*AuthenticationContext, error)
}

type AuthenticationContext struct {
	UserId int
}

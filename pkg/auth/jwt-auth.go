package myauth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	jwtTokenFieldName = "token"
)

type authClaims struct {
	AuthenticationContext
	jwt.StandardClaims
}

type JWTAuthentication struct {
	jwtKey []byte
}

func CreateJWTAuthentication(jwtKey []byte) *JWTAuthentication {
	return &JWTAuthentication{
		jwtKey: jwtKey,
	}
}

func (a *JWTAuthentication) LoginUser(w http.ResponseWriter, context AuthenticationContext) error {
	expirationTime := time.Now().Add(time.Minute * 10)

	claims := &authClaims{
		AuthenticationContext: context,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.jwtKey)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:    jwtTokenFieldName,
		Value:   tokenString,
		Expires: expirationTime,
	})

	return nil
}

func (a *JWTAuthentication) Authenticate(w http.ResponseWriter, r *http.Request) (*AuthenticationContext, error) {
	claims, err := a.getClaimsWithTokenCheck(w, r)

	if err != nil {
		return nil, err
	}

	return &claims.AuthenticationContext, nil
}

func (a *JWTAuthentication) getClaimsWithTokenCheck(w http.ResponseWriter, r *http.Request) (*authClaims, error) {
	cookie, err := r.Cookie(jwtTokenFieldName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	tokenStr := cookie.Value
	claims := &authClaims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return a.jwtKey, nil
		},
	)

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, errors.New("token is invalid")
	}

	return claims, nil
}

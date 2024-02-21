module github.com/roiciap/golang/user-api

go 1.21.4

require (
	github.com/roiciap/golang/myauth v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.19.0
	gopkg.in/validator.v2 v2.0.1
)

require github.com/golang-jwt/jwt v3.2.2+incompatible // indirect

replace github.com/roiciap/golang/myauth => ../myauth

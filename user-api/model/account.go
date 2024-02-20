package model

type Credentials struct {
	Login    string `json:"login" validate:"min=1"`
	Password string `json:"password" validate:"min=1"`
}

type Account struct {
	Login        string
	PasswordHash []byte
}

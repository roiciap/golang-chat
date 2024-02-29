package domain

type AccountDomain struct {
	ID           uint
	Login        string
	PasswordHash []byte
}

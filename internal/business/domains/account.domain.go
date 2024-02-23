package domainmodels

type AccountDomain struct {
	Login        string
	PasswordHash []byte
}

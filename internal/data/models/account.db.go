package dbmodels

import (
	"github.com/roiciap/golang-chat/internal/business/domain"
	"gorm.io/gorm"
)

type AccountDb struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string `gorm:"column:password_hash"`
}

func (acc *AccountDb) ToDomain() *domain.AccountDomain {
	return &domain.AccountDomain{
		ID:           acc.ID,
		Login:        acc.Username,
		PasswordHash: []byte(acc.PasswordHash),
	}
}

type AccountDbCreate struct {
	Login        string
	PasswordHash []byte
}

func (ac *AccountDbCreate) GetDbModel() AccountDb {
	return AccountDb{
		Username:     ac.Login,
		PasswordHash: string(ac.PasswordHash),
	}
}

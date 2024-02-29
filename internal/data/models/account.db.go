package dbmodels

import (
	"github.com/roiciap/golang-chat/internal/business/domain"
	dto "github.com/roiciap/golang-chat/internal/http/datatransfers/requests"
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

func AccountDbFromDto(data dto.AccountDto) AccountDb {
	return AccountDb{
		Username:     data.Login,
		PasswordHash: string(data.PasswordHash),
	}
}

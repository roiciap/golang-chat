package dbmodels

import (
	"gorm.io/gorm"
)

type AccountDb struct {
	gorm.Model
	Username     string
	PasswordHash string `gorm:"column:password_hash"`
}

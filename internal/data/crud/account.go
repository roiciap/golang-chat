package crud

import (
	"github.com/roiciap/golang-chat/internal/business/domain"
	dbmodels "github.com/roiciap/golang-chat/internal/data/models"
	dbservices "github.com/roiciap/golang-chat/internal/services/db"
	"gorm.io/gorm"
)

func AddAccount(toAdd dbmodels.AccountDbCreate) (uint, error) {
	db, err := getDb()
	if err != nil {
		return 0, err
	}

	dbData := toAdd.GetDbModel()
	result := db.Create(&dbData)

	if result.Error != nil {
		return 0, result.Error
	}

	return dbData.ID, nil
}

func GetAccountById(id uint) (*domain.AccountDomain, error) {
	record := &dbmodels.AccountDb{Model: gorm.Model{ID: id}}
	err := getAccount(record)
	if err != nil {
		return nil, err
	}

	acc := record.ToDomain()

	return acc, nil
}

func GetAccountByLogin(login string) (*domain.AccountDomain, error) {
	record := &dbmodels.AccountDb{Username: login}
	err := getAccount(record)
	if err != nil {
		return nil, err
	}

	acc := record.ToDomain()
	return acc, nil
}

func getAccount(acc *dbmodels.AccountDb) error {
	db, err := getDb()
	if err != nil {
		return err
	}
	result := db.Where(*acc).First(&acc)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func getDb() (*gorm.DB, error) {
	db, err := (&dbservices.AccountDatabaseSingleton{}).GetPostgresDB()
	return db, err
}

package migrations

import (
	dbmodels "github.com/roiciap/golang-chat/internal/data/models"
	dbservices "github.com/roiciap/golang-chat/internal/data/services"
)

func Up_1() error {
	db, err := (&dbservices.AccountDatabaseSingleton{}).GetPostgresDB()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	tables := []interface{}{}

	tables = addTable(db, &dbmodels.AccountDb{}, tables)

	err = db.Migrator().CreateTable(tables...)
	return err
}

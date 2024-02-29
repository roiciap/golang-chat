package dbservices

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AccountDatabaseSingleton struct {
	database *gorm.DB
}

func (s *AccountDatabaseSingleton) GetPostgresDB() (*gorm.DB, error) {
	var err error
	if s.database == nil {
		s.database, err = newGormDb()
	}
	return s.database, err
}

func newGormDb() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=admin dbname=account port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db, err
}

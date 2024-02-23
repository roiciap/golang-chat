package db

import "github.com/roiciap/golang-chat/internal/data/db/migrations"

func Migrate() error {
	return migrations.Up_1()
}

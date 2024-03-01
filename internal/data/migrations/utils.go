package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

func addTable(database *gorm.DB, model interface{}, tables []interface{}) []interface{} {
	if !database.Migrator().HasTable(model) {
		tables = append(tables, model)
	} else {
		fmt.Println("Tabela juz jest")
	}
	return tables
}

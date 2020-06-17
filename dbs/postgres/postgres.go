package postgres

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Use postgres.DB() for accessing records in your service
func DB() (*gorm.DB, error) {
	if instance.connection == nil {
		return nil, ErrNotConnected
	}
	return instance.connection, nil
}

//Use postgres.AddMigration() for all schema migrations in your  service within  "Service Interface"
func AddMigration(id string, m migration) {
	migrations = append(migrations, m)
}

func AddColumn(table Table, col *Column) *addColumnMigration {
	m := &addColumnMigration{
		tableName: table.Name,
		column:    col,
	}
	m.condition = &ifColumnNotExistsCondition{
		tableName:  table.Name,
		columnName: col.Name}
	return m
}

func AddIndex(table Table, index *Index) *addIndexMigration {
	m := &addIndexMigration{
		tableName: table.Name,
		index:     index,
	}
	m.condition = &ifIndexNotExistsCondition{
		tableName: table.Name,
		indexName: index.XName(table.Name),
	}
	return m
}

func DropIndex(table Table, index *Index) *dropIndexMigration {
	m := &dropIndexMigration{
		tableName: table.Name,
		index:     index,
	}
	m.condition = &ifIndexExistsCondition{
		tableName: table.Name,
		indexName: index.XName(table.Name),
	}
	return m
}

func AddTable(table Table) *addTableMigration {
	for _, col := range table.Columns {
		if col.IsPrimaryKey {
			table.PrimaryKeys = append(table.PrimaryKeys, col.Name)
		}
	}
	return &addTableMigration{
		table: table,
	}
}

func DropTable(tableName string) *dropTableMigration {
	return &dropTableMigration{
		tableName: tableName,
	}
}

func RenameTable(oldName string, newName string) *renameTableMigration {
	return &renameTableMigration{
		oldName: oldName,
		newName: newName,
	}
}

func CopyTableData(targetTable string, sourceTable string, colMap map[string]string) *copyTableDataMigration {
	m := &copyTableDataMigration{sourceTable: sourceTable, targetTable: targetTable}
	for key, value := range colMap {
		m.targetCols = append(m.targetCols, key)
		m.sourceCols = append(m.sourceCols, value)
	}
	return m
}

func TableCharset(tableName string, columns []*Column) *tableCharsetMigration {
	return &tableCharsetMigration{
		tableName: tableName,
		columns:   columns,
	}
}

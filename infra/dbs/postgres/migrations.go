package postgres

import (
	"fmt"
	"strings"
)

var (
	noOpSql = "SELECT 0;"
)

type migration interface {
	Sql() string
	ID() string
	SetID(id string)
	GetCondition() migrationCondition
}

type migrationBase struct {
	id        string
	condition migrationCondition
}

type rawSqlMigration struct {
	migrationBase
	sql string
}

type addColumnMigration struct {
	migrationBase
	tableName string
	column    *Column
}

type addIndexMigration struct {
	migrationBase
	tableName string
	index     *Index
}

type dropIndexMigration struct {
	migrationBase
	tableName string
	index     *Index
}

type addTableMigration struct {
	migrationBase
	table Table
}

type dropTableMigration struct {
	migrationBase
	tableName string
}

type renameTableMigration struct {
	migrationBase
	oldName string
	newName string
}

type copyTableDataMigration struct {
	migrationBase
	sourceTable string
	targetTable string
	sourceCols  []string
	targetCols  []string
}

type tableCharsetMigration struct {
	migrationBase
	tableName string
	columns   []*Column
}

func (m *migrationBase) GetCondition() migrationCondition {
	return m.condition
}

func (m *migrationBase) ID() string {
	return m.id
}

func (m *migrationBase) SetID(id string) {
	m.id = id
}

func (m *rawSqlMigration) Sql() string {
	if m.sql != "" {
		return m.sql
	}
	return noOpSql
}

func (m *addColumnMigration) Table(tableName string) *addColumnMigration {
	m.tableName = tableName
	return m
}

func (m *addColumnMigration) Column(col *Column) *addColumnMigration {
	m.column = col
	return m
}

func (m *addColumnMigration) Sql() string {
	return fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN %s", m.tableName, m.column.StringNoPk())
}

func (m *addIndexMigration) Table(tableName string) *addIndexMigration {
	m.tableName = tableName
	return m
}

func (m *addIndexMigration) Sql() string {
	var unique string
	var idxName string
	if m.index.Type == UniqueIndex {
		unique = " UNIQUE"
	}
	idxName = m.index.XName(m.tableName)
	return fmt.Sprintf("CREATE%s INDEX %v ON %v (%v)", unique, "\""+idxName+"\"", "\""+m.tableName+"\"", "\""+strings.Join(m.index.Cols, "\",\"")+"\"")
}

func (m *dropIndexMigration) Sql() string {
	if m.index.Name == "" {
		m.index.Name = strings.Join(m.index.Cols, "_")
	}
	idxName := m.index.XName(m.tableName)
	return fmt.Sprintf("DROP INDEX \"%v\" CASCADE", idxName)
}

func (m *addTableMigration) Sql() string {
	sql := "CREATE TABLE IF NOT EXISTS "
	sql += "\"" + m.table.Name + "\" (\n"
	pkList := m.table.PrimaryKeys

	for _, col := range m.table.Columns {
		if col.IsPrimaryKey && len(pkList) == 1 {
			sql += col.ColString()
		} else {
			sql += col.StringNoPk()
		}
		sql = strings.TrimSpace(sql)
		sql += "\n, "
	}

	if len(pkList) > 1 {
		quotedCols := []string{}
		for _, col := range pkList {
			quotedCols = append(quotedCols, "\""+col+"\"")
		}
		sql += "PRIMARY KEY ( " + strings.Join(quotedCols, ",") + " ), "
	}
	sql = sql[:len(sql)-2] + ");"
	return sql
}

func (m *dropTableMigration) Sql() string {
	return fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", m.tableName)
}

func (m *renameTableMigration) Rename(oldName string, newName string) *renameTableMigration {
	m.oldName = oldName
	m.newName = newName
	return m
}

func (m *renameTableMigration) Sql() string {
	return fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\"", m.oldName, m.newName)
}

func (m *copyTableDataMigration) Sql() string {
	sourceColsSql := quoteColList(m.sourceCols)
	targetColsSql := quoteColList(m.targetCols)

	return fmt.Sprintf("INSERT INTO \"%s\" (%s) SELECT %s FROM \"%s\"", m.targetTable, targetColsSql, sourceColsSql, m.sourceTable)
}

func quoteColList(cols []string) string {
	var sourceColsSql = ""
	for _, col := range cols {
		sourceColsSql += "\"" + col + "\""
		sourceColsSql += "\n, "
	}
	return strings.TrimSuffix(sourceColsSql, "\n, ")
}

func (m *tableCharsetMigration) Sql() string {
	var statements = []string{}
	for _, col := range m.columns {
		statements = append(statements, "ALTER \""+col.Name+"\" TYPE "+col.SqlType())
	}
	return "ALTER TABLE \"" + m.tableName + "\" " + strings.Join(statements, ", ") + ";"
}

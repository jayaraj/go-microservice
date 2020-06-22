package postgres

import "database/sql"

type migrationCondition interface {
	Sql() (string, []interface{})
	IsFulfilled(rows *sql.Rows, count int64) bool
}
type existsMigrationCondition struct {
}
type notExistsMigrationCondition struct {
}
type ifIndexExistsCondition struct {
	existsMigrationCondition
	tableName string
	indexName string
}
type ifIndexNotExistsCondition struct {
	notExistsMigrationCondition
	tableName string
	indexName string
}
type ifColumnNotExistsCondition struct {
	notExistsMigrationCondition
	tableName  string
	columnName string
}

func (c *existsMigrationCondition) IsFulfilled(rows *sql.Rows, count int64) bool {
	return count >= 1
}

func (c *notExistsMigrationCondition) IsFulfilled(rows *sql.Rows, count int64) bool {
	return count == 0
}

func (c *ifIndexExistsCondition) Sql() (string, []interface{}) {
	args := []interface{}{c.tableName, c.indexName}
	sql := "SELECT 1 FROM \"pg_indexes\" WHERE \"tablename\"=? AND \"indexname\"=?"
	return sql, args
}

func (c *ifIndexNotExistsCondition) Sql() (string, []interface{}) {
	args := []interface{}{c.tableName, c.indexName}
	sql := "SELECT 1 FROM \"pg_indexes\" WHERE \"tablename\"=? AND \"indexname\"=?"
	return sql, args
}

func (c *ifColumnNotExistsCondition) Sql() (string, []interface{}) {
	return "", nil
}

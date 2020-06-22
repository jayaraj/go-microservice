package postgres

import (
	"fmt"
	"strconv"
	"strings"
)

type Table struct {
	Name        string
	Columns     []*Column
	PrimaryKeys []string
	Indices     []*Index
}

type Column struct {
	Name            string
	Type            string
	Length          int
	Nullable        bool
	IsPrimaryKey    bool
	IsAutoIncrement bool
	Default         string
}

const (
	IndexType = iota + 1
	UniqueIndex
)

type Index struct {
	Name string
	Type int
	Cols []string
}

var (
	DB_Bit      = "BIT"
	DB_SmallInt = "SMALLINT"
	DB_Integer  = "INTEGER"
	DB_BigInt   = "BIGINT"

	DB_Enum = "ENUM"
	DB_Set  = "SET"

	DB_Char    = "CHAR"
	DB_Varchar = "VARCHAR"
	DB_Text    = "TEXT"
	DB_Uuid    = "UUID"

	DB_Date       = "DATE"
	DB_Time       = "TIME"
	DB_TimeStamp  = "TIMESTAMP"
	DB_TimeStampz = "TIMESTAMP WITH TIME ZONE"

	DB_Decimal = "DECIMAL"
	DB_Numeric = "NUMERIC"

	DB_Real   = "REAL"
	DB_Double = "DOUBLE PRECISION"

	DB_Bytea = "BYTEA"

	DB_Bool = "BOOL"

	DB_Serial    = "SERIAL"
	DB_BigSerial = "BIGSERIAL"

	DB_Hstore = "HSTORE"
)

func (column *Column) SqlType() string {
	if column.IsAutoIncrement {
		return DB_Serial
	}
	if column.Type == DB_Serial || column.Type == DB_BigSerial {
		column.IsAutoIncrement = true
		column.Nullable = false
	}
	resp := column.Type
	if column.Length > 0 {
		resp += "(" + strconv.Itoa(column.Length) + ")"
		return resp
	}
	return resp
}

func (column *Column) StringNoPk() string {
	sql := "\"" + column.Name + "\" "
	sql += column.SqlType() + " "
	if column.Nullable {
		sql += "NULL "
	} else {
		sql += "NOT NULL "
	}
	if column.Default != "" {
		sql += "DEFAULT " + column.DefaultValue() + " "
	}
	return sql
}

func (column *Column) ColString() string {
	sql := "\"" + column.Name + "\" "
	sql += column.SqlType() + " "

	if column.IsPrimaryKey {
		sql += "PRIMARY KEY "
	}
	if column.Nullable {
		sql += "NULL "
	} else {
		sql += "NOT NULL "
	}
	if column.Default != "" {
		sql += "DEFAULT " + column.DefaultValue() + " "
	}
	return sql
}

func (column *Column) DefaultValue() string {
	if column.Type == DB_Bool {
		if column.Default == "0" {
			return "FALSE"
		}
		return "TRUE"
	}
	return column.Default
}

func (index *Index) XName(tableName string) string {
	if index.Name == "" {
		index.Name = strings.Join(index.Cols, "_")
	}

	if !strings.HasPrefix(index.Name, "UQE_") &&
		!strings.HasPrefix(index.Name, "IDX_") {
		if index.Type == UniqueIndex {
			return fmt.Sprintf("UQE_%v_%v", tableName, index.Name)
		}
		return fmt.Sprintf("IDX_%v_%v", tableName, index.Name)
	}
	return index.Name
}

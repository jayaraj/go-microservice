package postgres

import (
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var (
	migrations []migration
)

type MigrationLog struct {
	Id          int64     `json:"id"`
	MigrationId string    `json:"migration_id"`
	Sql         string    `json:"sql"`
	Success     bool      `json:"success"`
	Error       string    `json:"error"`
	Timestamp   time.Time `json:"timestamp"`
}

func init() {
	migrations = make([]migration, 0)
	addMigrationLogMigrations()
}

func getMigrationLog() (map[string]MigrationLog, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	logMap := make(map[string]MigrationLog)
	logItems := make([]MigrationLog, 0)
	if !db.HasTable(new(MigrationLog)) {
		return logMap, nil
	}
	if err = db.Find(&logItems).Error; err != nil {
		return nil, err
	}
	for _, logItem := range logItems {
		if !logItem.Success {
			continue
		}
		logMap[logItem.MigrationId] = logItem
	}
	return logMap, nil
}

func startMigrations() error {
	db, err := DB()
	if err != nil {
		return err
	}

	logMap, err := getMigrationLog()
	if err != nil {
		return err
	}

	for _, m := range migrations {
		_, exists := logMap[m.ID()]
		if exists {
			log.WithField("ID", m.ID()).Debug("Skipping migration, already executed")
			continue
		}
		sql := m.Sql()
		record := MigrationLog{
			MigrationId: m.ID(),
			Sql:         sql,
			Timestamp:   time.Now(),
		}

		db.Transaction(func(tx *gorm.DB) error {
			err := executeMigration(m, tx)
			if err != nil {
				record.Error = err.Error()
			} else {
				record.Success = true
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			return nil
		})
	}
	return nil
}

func executeMigration(m migration, tx *gorm.DB) error {
	log.WithField("ID", m.ID()).Info("Executing migration")

	condition := m.GetCondition()
	if condition != nil {
		sql, args := condition.Sql()

		if sql != "" {
			log.WithFields(log.Fields{
				"ID":   m.ID(),
				"SQL":  sql,
				"Args": args,
			}).Debug("Executing migration condition")
			var count int64
			rows, err := tx.Raw(sql, args...).Count(&count).Rows()
			defer rows.Close()
			if err != nil {
				log.WithFields(log.Fields{
					"ID":    m.ID(),
					"Error": err,
				}).Error("Executing migration condition failed")
				return err
			}

			if !condition.IsFulfilled(rows, count) {
				log.WithFields(log.Fields{
					"ID":    m.ID(),
					"Error": err,
				}).Warn("Skipping migration already executed, but not recorded in migration log")
				return nil
			}
		}
	}

	if err := tx.Exec(m.Sql()).Error; err != nil {
		log.WithFields(log.Fields{
			"ID":    m.ID(),
			"Error": err,
		}).Error("Executing migration failed")
		return err
	}
	return nil
}

func addMigrationLogMigrations() {
	migrationLogV1 := Table{
		Name: "migration_log",
		Columns: []*Column{
			{Name: "id", Type: DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "migration_id", Type: DB_Varchar, Length: 255},
			{Name: "sql", Type: DB_Text},
			{Name: "success", Type: DB_Bool},
			{Name: "error", Type: DB_Text},
			{Name: "timestamp", Type: DB_TimeStamp},
		},
	}
	AddMigration("create migration_log table", AddTable(migrationLogV1))
}

package postgres

import (
	"errors"
	"fmt"
	"go-microservice/server"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBname   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Sslmode  string `json:"sslmode"`
}

type postgres struct {
	connection *gorm.DB
	config     config
}

var (
	instance         *postgres
	ErrNotConfigured = errors.New("Postgres is not configured")
	ErrNotConnected  = errors.New("Postgres is not connected")
)

func connect() {
	if err := instance.connect(); err != nil {
		log.WithField("Error", err).Errorln("Postgres connection failed")
		go func() {
			time.Sleep(30 * time.Second)
			go connect()
		}()
		return
	}
	log.Info("Postgres connected starting migrations")
	if err := startMigrations(); err != nil {
		log.WithField("error", err).Fatal("Migration failed")
	}
}

func init() {
	instance = &postgres{
		connection: nil,
	}
	//Set it High or low based on your requirement
	server.RegisterService(instance, server.Low)
}

func (c *postgres) Init() (err error) {
	if viper.IsSet("postgres") {
		config := &config{}
		if err := viper.UnmarshalKey("postgres", config); err != nil {
			return err
		}
		c.config = *config
		go connect()
		return nil
	}
	return ErrNotConfigured
}

func (c *postgres) connect() error {
	var connection *gorm.DB
	var err error

	args := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		c.config.Host, c.config.Port, c.config.Username, c.config.DBname, c.config.Password, c.config.Sslmode)
	if connection, err = gorm.Open("postgres", args); err != nil {
		return err
	}
	if viper.Get("mode") == "prod" {
		connection.LogMode(false)
	} else {
		connection.LogMode(true)
	}
	c.connection = connection
	c.connection.SingularTable(true)
	return nil
}

func (c *postgres) OnConfig() {
	//Do nothing
}

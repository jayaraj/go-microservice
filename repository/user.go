package repository

import (
	"go-microservice/dtos"
	"go-microservice/infra/bus"
	"go-microservice/infra/cache"
	"go-microservice/infra/dbs/postgres"
	"go-microservice/infra/server"
	"time"

	"github.com/jinzhu/gorm"
)

type userRepo struct{}

func init() {
	server.RegisterService(&userRepo{}, server.Low)
}

func (c *userRepo) Init() (err error) {

	c.addUserMigrations()

	//Register for all the repository requests
	bus.AddHandler(CreateUser)
	bus.AddHandler(ListUsers)
	return nil
}

func (c *userRepo) OnConfig() {
}

func (c *userRepo) addUserMigrations() {
	userV1 := postgres.Table{
		Name: "user",
		Columns: []*postgres.Column{
			{Name: "id", Type: postgres.DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "name", Type: postgres.DB_Varchar, Length: 255},
			{Name: "created", Type: postgres.DB_TimeStamp},
			{Name: "updated", Type: postgres.DB_TimeStamp},
		},
	}
	postgres.AddMigration("create user table", postgres.AddTable(userV1))

	//eg. on CR123456
	postgres.AddMigration("CR123456 set timezone", postgres.RawSql("SET timezone=UTC"))

	//eg. on CR987654
	postgres.AddMigration("CR987654 add email to user", postgres.AddColumn(userV1, &postgres.Column{
		Name: "email", Type: postgres.DB_Varchar, Length: 255, Nullable: true,
	}))
}

func CreateUser(cmd *dtos.CreateUserCmd) error {
	db, err := postgres.DB()
	if err != nil {
		return err
	}
	go cache.Delete(false, "userscount")
	return db.Transaction(func(tx *gorm.DB) error {
		user := dtos.User{
			Name:    cmd.Name,
			Email:   cmd.Email,
			Created: time.Now(),
			Updated: time.Now(),
		}
		err := tx.Create(&user).Error
		if err == nil {
			cmd.Result = &user
		}
		return err
	})
	return nil
}

func ListUsers(cmd *dtos.ListUsersCmd) error {
	var userCount int64
	userCount = 0
	db, err := postgres.DB()
	if err != nil {
		return err
	}
	err = cache.Get(false, "userscount", &userCount)
	if err != nil {
		db.Table("user").Count(&userCount)
		go cache.Set(false, "userscount", userCount, cache.ForEverNeverExpiry)
	}
	cmd.Result.Users = make([]*dtos.User, 0)
	if cmd.Limit*(cmd.Page-1) < userCount {
		err := db.Offset(cmd.Limit * (cmd.Page - 1)).Limit(cmd.Limit).Find(&cmd.Result.Users).Error
		return err
	}
	return nil
}

package dtos

import "time"

type User struct {
	Id      int64     `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Updated time.Time `json:"updated"`
	Created time.Time `json:"created"`
}

type CreateUserCmd struct {
	Name   string
	Email  string
	Result *User
}

type ListUsersCmd struct {
	Limit  int64
	Page   int64
	Result UsersResult
}

type UsersResult struct {
	Users []*User `json:"users"`
}

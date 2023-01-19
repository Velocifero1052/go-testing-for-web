package repository

import (
	"database/sql"
	"testingCourserWeb/pkg/data"
)

type DatabaseRepo interface {
	Connection() *sql.DB
	AllUsers() ([]*data.User, error)
	GetUser(id int) (*data.User, error)
	GetUserByEmail(email string) (*data.User, error)
	UpdateUser(u data.User) error
	DeleteUser(id int) error
	ResetPassword(id int, password string) error
	InsertUserImage(i data.UserImage) (int, error)
}

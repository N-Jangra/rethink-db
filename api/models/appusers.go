package models

import (
	"time"

	rethinkdb "github.com/rethinkdb/rethinkdb-go"
)

type rdbSess struct {
	session   *rethinkdb.Session
	dbName    string
	tableName string
}

type AppUser struct {
	Userid    string    ` json:"userid" rethinkdb:"userid" `
	Name      string    ` json:"name" rethinkdb:"name" `
	Details   string    ` json:"details,omitempty" rethink:"details" `
	Password  string    ` json:"password" rethinkdb:"password" `
	Active    bool      ` json:"active" rethinkdb:"active" `
	Role      string    ` json:"role" rethinkdb:"role" `
	Dob       time.Time ` json:"dob" rethinkdb:"dob" `
	Gender    string    ` json:"gender" rethinkdb:"gender" `
	Email     string    ` json:"email" rethinkdb:"email" `
	Phone     string    ` json:"phone" rethinkdb:"phone" `
	CreatedAt time.Time ` json:"createdat" rethinkdb:"createdat" `
	UpdatedAt time.Time ` json:"updatedat" rethinkdb:"updatedat" `
}

package db_bcrypt

import (
	"code.google.com/p/go.crypto/bcrypt"
	"database/sql"
	"fmt"
	"github.com/lye/crud"
	"macrobooru/api/pluggable"
	"macrobooru/models"
)

type userStore struct {
	pluggable.UserStore

	q string
}

func init() {
	pluggable.RegisterUserStore(newUserStore())
}

func newUserStore() *userStore {
	return &userStore{
		q: fmt.Sprintf("SELECT * FROM %s WHERE username = $1", models.UserMeta.TableName()),
	}
}

func (store *userStore) SetUserPassword(db *sql.DB, username, password string) (bool, error) {
	rows, er := db.Query(store.q, username)
	if er != nil {
		return true, er
	}
	defer rows.Close()

	if !rows.Next() {
		return false, nil
	}

	var u models.User

	if er := crud.Scan(rows, &u); er != nil {
		return true, er
	}

	hashBytes, er := bcrypt.GenerateFromPassword([]byte(password), 10)
	if er != nil {
		return true, er
	}

	u.Passhash = string(hashBytes)

	if er := crud.Update(db, models.UserMeta.TableName(), models.UserMeta.PrimaryKey(), u); er != nil {
		return true, er
	}

	return true, nil
}

func (store *userStore) Authenticate(db *sql.DB, username, password string) (*models.User, error) {
	rows, er := db.Query(store.q, username)
	if er != nil {
		return nil, er
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var u models.User

	if er := crud.Scan(rows, &u); er != nil {
		return nil, er
	}

	if er := bcrypt.CompareHashAndPassword([]byte(u.Passhash), []byte(password)); er != nil {
		if er == bcrypt.ErrMismatchedHashAndPassword {
			return nil, nil

		} else {
			return nil, er
		}
	}

	return &u, nil
}

package pluggable

import (
	"database/sql"
	"macrobooru/models"
	"sync"
)

type UserStore interface {
	Authenticate(db *sql.DB, username, password string) (*models.User, error)
	SetPassword(db *sql.DB, username, password string) (userExists bool, er error)
}

var userStoreLock sync.RWMutex
var userStores []UserStore

func RegisterUserStore(store UserStore) {
	userStoreLock.Lock()
	defer userStoreLock.Unlock()

	userStores = append(userStores, store)
}

func CheckCredentials(db *sql.DB, username, password string) (*models.User, error) {
	userStoreLock.RLock()
	defer userStoreLock.RUnlock()

	for _, store := range userStores {
		if u, er := store.Authenticate(db, username, password); er != nil {
			return nil, er

		} else if u != nil {
			return u, nil
		}
	}

	return nil, nil
}

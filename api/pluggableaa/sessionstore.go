package pluggable

import (
	"fmt"
	"macrobooru/models"
	"sync"
)

type SessionKey string

type SessionStore interface {
	CreateSession(userGuid models.GUID) (SessionKey, error)
	DestroySession(key SessionKey) error
	LookupSession(key SessionKey) (*models.GUID, error)
}

var sessionStoreLock sync.RWMutex
var sessionStores []SessionStore

func RegisterSessionStore(store SessionStore) {
	sessionStoreLock.Lock()
	defer sessionStoreLock.Unlock()

	sessionStores = append(sessionStores, store)
}

func CreateSession(userGuid models.GUID) (key SessionKey, er error) {
	sessionStoreLock.RLock()
	defer sessionStoreLock.RUnlock()

	for _, store := range sessionStores {
		if key, er = store.CreateSession(userGuid); er == nil && key != "" {
			return key, nil
		}
	}

	if er == nil {
		er = fmt.Errorf("No session stores currently registered")
	}

	return
}

func DestroySession(key SessionKey) {
	sessionStoreLock.RLock()
	defer sessionStoreLock.RUnlock()

	for _, store := range sessionStores {
		/* Shouldn't really ignore an error here, but meh */
		store.DestroySession(key)
	}
}

func LookupSession(key SessionKey) (*models.GUID, error) {
	sessionStoreLock.RLock()
	defer sessionStoreLock.RUnlock()

	for _, store := range sessionStores {
		if userGuid, er := store.LookupSession(key); er != nil {
			return nil, er

		} else if userGuid != nil {
			return userGuid, nil
		}
	}

	return nil, nil
}

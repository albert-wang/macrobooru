package naive

import (
	"crypto/rand"
	"fmt"

	. "macrobooru/api/pluggable"
	"macrobooru/models"
)

type sessionStore struct {
	SessionStore

	sessions map[SessionKey]models.GUID
}

func init() {
	RegisterSessionStore(newSessionStore())
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		sessions: make(map[SessionKey]models.GUID),
	}
}

func (store *sessionStore) CreateSession(userGuid models.GUID) (SessionKey, error) {
	bs := make([]byte, 40)

	if _, er := rand.Read(bs); er != nil {
		return "", er
	}

	key := SessionKey(fmt.Sprintf("%x", bs))
	store.sessions[key] = userGuid

	return key, nil
}

func (store *sessionStore) DestroySession(key SessionKey) error {
	delete(store.sessions, key)
	return nil
}

func (store *sessionStore) LookupSession(key SessionKey) (*models.GUID, error) {
	if userGuid, ok := store.sessions[key]; ok {
		return &userGuid, nil
	}

	return nil, nil
}

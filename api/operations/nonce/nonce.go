package nonce

import (
	"database/sql"
	"encoding/json"

	"macrobooru/api"
	"macrobooru/models"
)

/* Takes no arguments, but has internal state */
type NoncePayload struct {
	Email string `json:"email"`
}

func (np *NoncePayload) Name() string {
	return "nonce"
}

func (np *NoncePayload) Parse(req *api.RequestWrapper) (api.Operation, error) {
	var payload NoncePayload

	if er := json.Unmarshal(req.RawData, &payload); er != nil {
		return nil, er
	}

	return &payload, nil
}

func (ap *NoncePayload) ParseResponse(resWrapper api.ResponseWrapper) (interface{}, error) {
	return nil, nil
}

func (np *NoncePayload) Process(u *models.User, db *sql.DB) (interface{}, error) {
	return nil, nil
}

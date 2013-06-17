package resetpassword

import (
	"database/sql"
	"encoding/json"

	"macrobooru/api"
	"macrobooru/models"
)

type ResetPasswordPayload struct {
	Username string `json:"username, omitempty"`
	Email    string `json:"email, omitempty"`
}

func (rpp *ResetPasswordPayload) Name() string {
	return "resetpassword"
}

func (rpp *ResetPasswordPayload) Parse(req *api.RequestWrapper) (api.Operation, error) {
	var payload ResetPasswordPayload

	if er := json.Unmarshal(req.RawData, &payload); er != nil {
		return nil, er
	}

	return &payload, nil
}

func (ap *ResetPasswordPayload) ParseResponse(resWrapper api.ResponseWrapper) (interface{}, error) {
	return nil, nil
}

func (rpp *ResetPasswordPayload) Process(u *models.User, db *sql.DB) (interface{}, error) {
	return nil, nil
}

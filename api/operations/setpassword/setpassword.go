package setpassword

import (
	"database/sql"
	"encoding/json"

	"macrobooru/api"
	"macrobooru/models"
)

type SetPasswordPayload struct {
	// OriginalPassword is the user's original password.
	OriginalPassword string `json:"original,omitempty"`

	// Target is the GUID of the User this password set changes.
	Target models.GUID `json:"target,omitempty"`

	// Next is the value to set as the User's new password.
	Next string `json:"next"`

	//Optional nonce to avoid setting OriginalPassword and Target.
	Nonce string `json:"nonce,omitempty"`
}

func (spp *SetPasswordPayload) Name() string {
	return "setpassword"
}

func (spp *SetPasswordPayload) Parse(req *api.RequestWrapper) (api.Operation, error) {
	var payload SetPasswordPayload

	if er := json.Unmarshal(req.RawData, &payload); er != nil {
		return nil, er
	}

	return &payload, nil
}

func (ap *SetPasswordPayload) ParseResponse(resWrapper api.ResponseWrapper) (interface{}, error) {
	return nil, nil
}

func (spp *SetPasswordPayload) Process(u *models.User, db *sql.DB) (interface{}, error) {
	return nil, nil
}

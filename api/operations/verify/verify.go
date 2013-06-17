package verify

import (
	"database/sql"
	"encoding/json"
	"macrobooru/api"
	"macrobooru/models"
)

type VerifyPayload struct {
	Code string `json:"code"`
}

func (vp *VerifyPayload) Name() string {
	return "verify"
}

func (vp *VerifyPayload) Parse(req *api.RequestWrapper) (api.Operation, error) {
	var payload VerifyPayload

	if er := json.Unmarshal(req.RawData, &payload); er != nil {
		return nil, er
	}

	return &payload, nil
}

func (vp *VerifyPayload) Process(u *models.User, db *sql.DB) (interface{}, error) {
	return nil, nil
}

func (vp *VerifyPayload) ParseResponse(req api.ResponseWrapper) (interface{}, error) {
	res := VerifyResponse{}
	er := json.Unmarshal(req.Data, &res)
	return &res, er
}

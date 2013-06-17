package static_status

import (
	"database/sql"
	"macrobooru/api"
	"macrobooru/models"

	"encoding/json"
)

type StaticStatusPayload struct {
	// Target is the GUID of the User this password set changes.
	Ids []string `json:"ids,omitempty"`
}

func (spp *StaticStatusPayload) Name() string {
	return "static_status"
}

func (spp *StaticStatusPayload) Parse(req *api.RequestWrapper) (api.Operation, error) {
	var payload StaticStatusPayload

	if er := json.Unmarshal(req.RawData, &payload); er != nil {
		return nil, api.ErrorInvalidInputFormat(er.Error())
	}

	return &payload, nil
}

func (ap *StaticStatusPayload) ParseResponse(resWrapper api.ResponseWrapper) (interface{}, error) {
	return nil, nil
}

func (spp *StaticStatusPayload) Process(u *models.User, db *sql.DB) (interface{}, error) {
	return nil, nil
}

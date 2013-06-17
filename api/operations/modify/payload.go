package modify

import (
	"macrobooru/api"
	"macrobooru/models"

	"database/sql"
	"encoding/json"
)

type ModifyPayload []ModifyRequest

func (*ModifyPayload) Name() string {
	return "modify"
}

func (*ModifyPayload) Parse(req *api.RequestWrapper) (api.Operation, error) {
	var payload ModifyPayload

	if er := json.Unmarshal(req.RawData, &payload); er != nil {
		return nil, api.ErrorInvalidInputFormat(er.Error())
	}

	return &payload, nil
}

func (payload *ModifyPayload) Process(u *models.User, db *sql.DB) (interface{}, error) {
	slice := []ModifyRequest(*payload)

	for idx := range slice {
		if er := slice[idx].verify(db); er != nil {
			return nil, er
		}
	}

	for idx := range slice {
		if er := slice[idx].run(db); er != nil {
			return nil, er
		}
	}

	return nil, nil
}

func (*ModifyPayload) ParseResponse(wrapper api.ResponseWrapper) (interface{}, error) {
	res := ModifyResponse{}

	if er := json.Unmarshal(wrapper.Data, &res); er != nil {
		return nil, api.ErrorGeneric(er)
	}

	return res, nil
}

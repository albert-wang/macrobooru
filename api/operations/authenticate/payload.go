package authenticate

import (
	"macrobooru/api"
	"macrobooru/models"

	"database/sql"
)

type AuthPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (ap *AuthPayload) Name() string {
	return "authenticate"
}

func (ap *AuthPayload) Parse(req *api.RequestWrapper) (api.Operation, error) {
	var payload AuthPayload

	return &payload, nil
}

func ParseResponse(resWrapper api.ResponseWrapper) (AuthResponse, error) {
	res := AuthResponse{}

	return res, nil
}

func (ap *AuthPayload) ParseResponse(resWrapper api.ResponseWrapper) (interface{}, error) {
	return ParseResponse(resWrapper)
}

func (ap *AuthPayload) Process(u *models.User, db *sql.DB) (interface{}, error) {
	return AuthResponse{
		Token: string(""),
	}, nil
}

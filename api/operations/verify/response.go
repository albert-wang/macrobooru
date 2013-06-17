package verify

import (
	"macrobooru/models"
)

type VerifyResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

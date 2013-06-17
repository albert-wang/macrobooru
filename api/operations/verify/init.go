package verify

import (
	"macrobooru/api"
)

func init() {
	api.RegisterOperation(&VerifyPayload{})
}

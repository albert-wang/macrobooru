package nonce

import (
	"macrobooru/api"
)

func init() {
	api.RegisterOperation(&NoncePayload{})
}

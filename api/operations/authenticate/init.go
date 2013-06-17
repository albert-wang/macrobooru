package authenticate

import (
	"macrobooru/api"
)

func init() {
	api.RegisterOperation(&AuthPayload{})
}

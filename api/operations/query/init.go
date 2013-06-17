package query

import (
	"macrobooru/api"
)

func init() {
	api.RegisterOperation(&QueryPayload{})
}

package modify

import (
	"macrobooru/api"
)

func init() {
	api.RegisterOperation(&ModifyPayload{})
}

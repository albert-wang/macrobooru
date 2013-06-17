package static_status

import (
	"macrobooru/api"
)

func init() {
	api.RegisterOperation(&StaticStatusPayload{})
}

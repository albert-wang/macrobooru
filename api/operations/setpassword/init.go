package setpassword

import (
	"macrobooru/api"
)

func init() {
	api.RegisterOperation(&SetPasswordPayload{})
}

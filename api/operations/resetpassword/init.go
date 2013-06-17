package resetpassword

import (
	"macrobooru/api"
)

func init() {
	api.RegisterOperation(&ResetPasswordPayload{})
}

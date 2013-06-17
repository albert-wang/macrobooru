package api

import (
	"encoding/json"
	"fmt"

	"macrobooru/models"
)

const (
	ErrCodeNoError                = 0x0
	ErrCodeGeneric                = 0x1
	ErrCodeInvalidInputFormat     = 0x2
	ErrCodeMissingInputField      = 0x3
	ErrCodeInvalidInputField      = 0x4
	ErrCodeObjectNotFound         = 0x5
	ErrCodeRequiresAuthentication = 0x6
	ErrCodeInvalidCredentials     = 0x7
	ErrCodeInvalidToken           = 0x8
	ErrCodeConversionFailure      = 0x9
	ErrCodeFileNotFound           = 0xA
	ErrCodeInvalidFileType        = 0xB
	ErrCodeVerificationEmailSent  = 0x80000000
	ErrCodeUsernameAlreadyExists  = 0x80000001
	ErrCodeEmailAlreadyExists     = 0x80000002
	ErrCodeAuthenticationFailure  = 0x80000003
	ErrCodeAccountNotVerified     = 0x80000004
	ErrCodeRequiresAdmin          = 0x80000005
	ErrCodeInvalidNonce           = 0x80000006
	ErrCodePermissionDenied       = 0x80000007
	ErrCodeAlreadyDeleted         = 0x80000008
	ErrCodeUserNotFound           = 0x80000009
	ErrCodeSphinxSyntaxError      = 0x8000000A
	ErrCodeSphinxOtherError       = 0x8000000B
)

type ApiError struct {
	code     int64
	args     []interface{}
	original error
}

func NewApiError(code int64, args []interface{}, original error) error {
	return &ApiError{
		code:     code,
		args:     args,
		original: original,
	}
}

func (e *ApiError) Error() string {
	var original string

	if e.original != nil {
		original = e.original.Error()
	}

	return fmt.Sprintf("ApiError{code: %d, args: %#v, original: %s}", e.code, e.args, original)
}

func (e *ApiError) Code() int64 {
	return e.code
}

func (e *ApiError) Original() (errorString string) {
	if e.original != nil {
		errorString = e.original.Error()
	} else {
		errorString = ""
	}

	return
}

func (e *ApiError) MarshalJSON() ([]byte, error) {
	if e == nil {
		return NoError().MarshalJSON()
	}
	var msg string
	if e.original != nil {
		msg = e.original.Error()
	}

	tmp := map[string]interface{}{
		"statusCode": e.code,
		"statusMsg":  msg,
		"data":       nil,
	}
	return json.Marshal(tmp)
}

func NoError() *ApiError {
	return &ApiError{ErrCodeNoError, nil, nil}
}

func ErrorGeneric(originals ...error) error {
	var original error = nil

	if len(originals) > 0 {
		original = originals[0]
	}

	return &ApiError{ErrCodeGeneric, nil, original}
}

func ErrorInvalidInputFormat(payload string) error {
	return &ApiError{ErrCodeInvalidInputFormat, []interface{}{payload}, nil}
}

func ErrorMissingInputField(field string) error {
	return &ApiError{ErrCodeMissingInputField, []interface{}{field}, nil}
}

func ErrorInvalidInputField(field string) error {
	return &ApiError{ErrCodeInvalidInputField, []interface{}{field}, nil}
}

func ErrorObjectNotFound(guid models.GUID) error {
	return &ApiError{ErrCodeObjectNotFound, []interface{}{guid}, nil}
}

func ErrorRequiresAuthentication() error {
	return &ApiError{ErrCodeRequiresAuthentication, nil, nil}
}

func ErrorInvalidCredentials() error {
	return &ApiError{ErrCodeInvalidCredentials, nil, nil}
}

func ErrorInvalidToken() error {
	return &ApiError{ErrCodeInvalidToken, nil, nil}
}

func ErrorConversionFailure() error {
	return &ApiError{ErrCodeConversionFailure, nil, nil}
}

func ErrorFileNotFound() error {
	return &ApiError{ErrCodeFileNotFound, nil, nil}
}

func ErrorInvalidFileType() error {
	return &ApiError{ErrCodeInvalidFileType, nil, nil}
}

func ErrorVerificationEmailSent() error {
	return &ApiError{ErrCodeVerificationEmailSent, nil, nil}
}

func ErrorUsernameAlreadyExists(username string) error {
	return &ApiError{ErrCodeUsernameAlreadyExists, []interface{}{username}, nil}
}

func ErrorEmailAlreadyExists(email string) error {
	return &ApiError{ErrCodeEmailAlreadyExists, []interface{}{email}, nil}
}

func ErrorAuthenticationFailure() error {
	return &ApiError{ErrCodeAuthenticationFailure, nil, nil}
}

func ErrorAccountNotVerified() error {
	return &ApiError{ErrCodeAccountNotVerified, nil, nil}
}

func ErrorInvalidNonce() error {
	return &ApiError{ErrCodeInvalidNonce, nil, nil}
}

func ErrorRequiresAdmin() error {
	return &ApiError{ErrCodeRequiresAdmin, nil, nil}
}

func ErrorPermissionDenied(unknown string) error {
	return &ApiError{ErrCodePermissionDenied, []interface{}{unknown}, nil}
}

func ErrorAlreadyDeleted(unknown string) error {
	return &ApiError{ErrCodeAlreadyDeleted, nil, nil}
}

func ErrorUserNotFound() error {
	return &ApiError{ErrCodeUserNotFound, nil, nil}
}

func ErrorSphinxSyntaxError(original error) error {
	return &ApiError{ErrCodeSphinxSyntaxError, nil, original}
}

func ErrorSphinxOtherError(original error) error {
	return &ApiError{ErrCodeSphinxOtherError, nil, original}
}

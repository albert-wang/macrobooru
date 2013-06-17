package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"macrobooru/api"
	"macrobooru/api/operations/authenticate"
	"macrobooru/api/operations/nonce"
	"macrobooru/api/operations/resetpassword"
	"macrobooru/api/operations/setpassword"
	_ "macrobooru/api/operations/static_status"
	"macrobooru/api/operations/verify"
	"macrobooru/models"
)

type Client struct {
	endpoint  string
	AuthToken string
	Config    ClientConfig
}

type ClientConfig struct {
	Debug                     bool `json:"debug"`
	VerifyEmails              bool `json:"verifyEmails"`
	DenyUnauthorizedAccess    bool `json:"denyUnauthorizedAccess"`
	RegistrationRequiresNonce bool `json:"registrationRequiresNonce"`
}

func NewClient(endpoint string) (*Client, error) {
	defaultClient := Client{
		endpoint: endpoint,
		Config: ClientConfig{
			Debug:                     false,
			VerifyEmails:              false,
			DenyUnauthorizedAccess:    false,
			RegistrationRequiresNonce: false,
		},
	}

	resp, er := http.Get(endpoint + "/config.js")
	if er != nil {
		return &defaultClient, nil
	}

	if resp.StatusCode != 200 {
		return &defaultClient, nil
	}

	defer resp.Body.Close()

	bodyBytes, er := ioutil.ReadAll(resp.Body)
	if er != nil {
		return nil, er
	}

	bodyStr := string(bodyBytes)

	// Read until the opening {
	for i := 0; i < len(bodyStr); i += 1 {
		if bodyStr[i] == '{' {
			bodyStr = bodyStr[i:]
			break
		}
	}

	// Cut the end until the closing }
	for i := len(bodyStr) - 1; i > 0; i -= 1 {
		if bodyStr[i] == '}' {
			bodyStr = bodyStr[:i+1]
			break
		}
	}

	var config ClientConfig
	if er := json.Unmarshal([]byte(bodyStr), &config); er != nil {
		log.Printf("Cannot unmarshal bytes:\n%s\n", bodyStr)
		return nil, er
	}

	return &Client{
		endpoint: endpoint,
		Config:   config,
	}, nil
}

func (client *Client) Authenticated() bool {
	return client.AuthToken != ""
}

func (client *Client) Authenticate(user, pass string) error {
	operation := &authenticate.AuthPayload{
		Username: user,
		Password: pass,
	}

	resWrapper, er := client.Execute(operation, nil)
	if er != nil {
		return er
	}

	result, er := operation.ParseResponse(*resWrapper)
	if er != nil {
		return er
	}

	if authResult, ok := result.(authenticate.AuthResponse); !ok {
		return fmt.Errorf("Unable to authenticate: result is a %T, not an authenticate.AuthResponse\n", result)

	} else {
		client.AuthToken = string(authResult.Token)
	}

	return nil
}

func (client *Client) Register(user, pass, email string, nonce string) (statusCode int64, er error) {
	payloadMap := map[string]interface{}{
		"username": user,
		"email":    email,
		"password": pass,
	}

	if nonce != "" {
		payloadMap["nonce"] = nonce
	}

	payloadJson, er := json.Marshal(payloadMap)
	if er != nil {
		return 0, er
	}

	buf := bytes.NewBuffer(payloadJson)
	body := ioutil.NopCloser(buf)

	req, er := http.NewRequest("POST", client.endpoint+"/register", body)
	if er != nil {
		return 0, er
	}

	req.ContentLength = int64(len(payloadJson))
	req.Header.Set("Content-Type", "application/json")

	resp, er := http.DefaultClient.Do(req)
	if er != nil {
		return 0, er
	}
	defer resp.Body.Close()

	respWrapper, er := api.UnwrapHttpResponse(resp.Body)
	if er != nil {
		return 0, er
	}

	return respWrapper.StatusCode, nil
}

func (client *Client) Verify(code string) (token string, user *models.User, er error) {
	payload := &verify.VerifyPayload{
		Code: code,
	}

	res, er := client.Execute(payload, nil)
	if er != nil {
		return
	}

	iRes, er := payload.ParseResponse(*res)
	if er != nil {
		return
	}

	verifyRes := iRes.(*verify.VerifyResponse)

	token = verifyRes.Token
	user = verifyRes.User
	return
}

func (client *Client) CreateNonce(email string) error {
	payload := &nonce.NoncePayload{
		Email: email,
	}

	_, er := client.Execute(payload, nil)
	if er != nil {
		return er
	}

	return nil
}

func (client *Client) RequestPasswordReset(email string) error {
	payload := &resetpassword.ResetPasswordPayload{
		Email:    email,
		Username: "",
	}

	_, er := client.Execute(payload, nil)
	if er != nil {
		return er
	}

	return nil
}

func (client *Client) SetPasswordWithNonce(pword string, nonce string) error {
	payload := &setpassword.SetPasswordPayload{
		Next:  pword,
		Nonce: nonce,
	}

	_, er := client.Execute(payload, nil)
	if er != nil {
		return er
	}

	return nil
}

func (client *Client) wrapRequest(operation api.Operation, attachments map[string]multipart.File) (*http.Request, error) {
	reqWrapper := api.RequestWrapper{
		Operation:   operation.Name(),
		AuthToken:   client.AuthToken,
		Data:        operation,
		Attachments: attachments,
	}

	httpReq, er := api.WrapHttpRequest(client.endpoint, &reqWrapper)
	if er != nil {
		return nil, er
	}

	return httpReq, nil
}

func (client *Client) dumpRequest(operation api.Operation, attachments map[string]multipart.File) {
	httpReq, er := client.wrapRequest(operation, attachments)
	if er == nil {
		body, _ := ioutil.ReadAll(httpReq.Body)
		fmt.Printf("=======\nREQUEST\n=======\n%s\n", string(body))
	}
}

func (client *Client) DumpQuery(query *Query) {
	payload := query.buildPayload()
	client.dumpRequest(&payload, nil)
}

func (client *Client) Execute(operation api.Operation, attachments map[string]multipart.File) (*api.ResponseWrapper, error) {
	httpReq, er := client.wrapRequest(operation, attachments)
	if er != nil {
		return nil, er
	}

	httpRes, er := http.DefaultClient.Do(httpReq)
	if er != nil {
		return nil, er
	}
	defer httpRes.Body.Close()

	resWrapper, er := api.UnwrapHttpResponse(httpRes.Body)
	if er != nil {
		return nil, er
	}

	if resWrapper.StatusCode != 0 {
		return nil, api.NewApiError(resWrapper.StatusCode, nil, fmt.Errorf("%s", resWrapper.StatusMessage))
	}

	return resWrapper, nil
}

func (client *Client) NewQuery() *Query {
	return NewQuery()
}

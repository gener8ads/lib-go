package recaptcha

import (
	"io/ioutil"
	"net/http"
	"net/url"

	jsoniter "github.com/json-iterator/go"
)

const endpoint = "https://www.google.com/recaptcha/api/siteverify"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// VerifyResponse from recaptcha
type VerifyResponse struct {
	Success  bool    `json:"success"`
	Score    float64 `json:"score"`
	Action   string  `json:"action"`
	Hostname string  `json:"hostname"`
}

// Verify a recaptcha token
func Verify(secret string, token string, ip string) (*VerifyResponse, error) {
	var res *VerifyResponse

	resp, reqErr := http.PostForm(endpoint, url.Values{
		"secret":   {secret},
		"response": {token},
		"remoteip": {ip},
	})

	if reqErr != nil {
		return res, reqErr
	}

	body, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return res, readErr
	}

	decodeErr := json.Unmarshal(body, &res)

	return res, decodeErr
}

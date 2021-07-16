package recaptcha

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gener8ads/lib-go/pkg/ginutil"
	"github.com/gener8ads/lib-go/pkg/jsonapi"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

const endpoint = "https://www.google.com/recaptcha/api/siteverify"

// CaptchaScore is a key, against which the score for the captcha is stored in the gin.Context, for later use
const CaptchaScore = "captcha_score"

const (
	versionHeader = "X-Captcha-Version"
	tokenHeader   = "X-Captcha-Token"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// VerifyResponse from recaptcha
type VerifyResponse struct {
	Success            bool      `json:"success"`
	Score              float64   `json:"score"`
	Action             string    `json:"action"`
	Hostname           string    `json:"hostname"`
	ChallengeTimestamp time.Time `json:"challenge_ts"`
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

// Middleware is the verification process implemented as a GIN middleware function
func Middleware(expectedAction string) gin.HandlerFunc {
	expectedAction = fmt.Sprintf("action/%s", expectedAction)

	enabled, _ := strconv.ParseBool(os.Getenv("RECAPTCHA_ENABLED"))

	secrets := map[int]string{
		2: os.Getenv("RECAPTCHA_V2_SECRET"),
		3: os.Getenv("RECAPTCHA_V3_SECRET"),
	}

	minScore, _ := strconv.ParseFloat(os.Getenv("RECAPTCHA_MIN_SCORE"), 32)

	expectedHostname := os.Getenv("RECAPTCHA_HOSTNAME")

	iMaxTime, _ := strconv.Atoi(os.Getenv("RECAPTCHA_MAX_TIME"))

	maxTime := time.Duration(iMaxTime)

	return func(c *gin.Context) {
		if !enabled {
			c.Set("log_field:captchaNotEnabled", "captcha not enabled")
			return
		}

		captchaVersion, err := strconv.Atoi(c.GetHeader(versionHeader))
		if err != nil {
			c.Set("log_field:captchaVersionErr", err.Error())
			jsonapi.Error(c, jsonapi.ErrorResponse{
				Status: http.StatusUnprocessableEntity,
				Code:   "recaptcha.missingVersion",
				Detail: "recaptcha.missingVersion",
			})
			return
		}

		secret, exists := secrets[captchaVersion]
		if !exists || secret == "" {
			c.Set("log_field:captchaSecretNotExists", fmt.Sprintf("%d", captchaVersion))
			jsonapi.Error(c, jsonapi.ErrorResponse{
				Status: http.StatusUnprocessableEntity,
				Code:   "recaptcha.missingVersion",
				Detail: "recaptcha.missingVersion",
			})
			return
		}

		res, err := Verify(secret, c.GetHeader(tokenHeader), ginutil.ExtractIP(c))
		if err != nil || !res.Success {
			if err != nil {
				c.Set("log_field:captchaVerifyErr", err.Error())
			} else {
				c.Set("log_field:captchaVerifyFailed", fmt.Sprintf("Success-%t~Score-%f~Action-%s~Hostname-%s~ChallengeTimestamp-%s", res.Success, res.Score, res.Action, res.Hostname, res.ChallengeTimestamp.Format(time.RFC3339Nano)))
			}
			jsonapi.Error(c, jsonapi.ErrorResponse{
				Status: http.StatusUnprocessableEntity,
				Code:   "recaptcha.failed",
				Detail: "recaptcha.failed",
			})
			return
		}

		if captchaVersion == 3 {
			if res.Action != expectedAction {
				c.Set("log_field:captchaUnexpectedAction", fmt.Sprintf("Expected-%s~Response-%s", expectedAction, res.Action))
				jsonapi.Error(c, jsonapi.ErrorResponse{
					Status: http.StatusUnprocessableEntity,
					Code:   "recaptcha.incorrectAction",
					Detail: "recaptcha.incorrectAction",
				})
				return
			}

			if res.Score < minScore {
				c.Set("log_field:captchaScoreTooLow", fmt.Sprintf("Minimum-%f~Response-%f", minScore, res.Score))
				jsonapi.Error(c, jsonapi.ErrorResponse{
					Status: http.StatusUnprocessableEntity,
					Code:   "recaptcha.challenge",
					Detail: "recaptcha.challenge",
				})
				return
			}
		}

		if expectedHostname != "" && res.Hostname != expectedHostname {
			c.Set("log_field:captchaUnexpectedHostname", fmt.Sprintf("Expected-%s~Response-%s", expectedHostname, res.Hostname))
			jsonapi.Error(c, jsonapi.ErrorResponse{
				Status: http.StatusUnprocessableEntity,
				Code:   "recaptcha.challenge",
				Detail: "recaptcha.challenge",
			})
			return
		}

		if maxTime > 0 && res.ChallengeTimestamp.Add(maxTime*time.Second).Before(time.Now()) {
			c.Set("log_field:captchaChallengeExpired", fmt.Sprintf("CurrentTime-%s~ChallengedExpiration-%s", time.Now().Format(time.RFC3339Nano), res.ChallengeTimestamp.Add(maxTime*time.Second).Format(time.RFC3339Nano)))
			jsonapi.Error(c, jsonapi.ErrorResponse{
				Status: http.StatusUnprocessableEntity,
				Code:   "recaptcha.challenge",
				Detail: "recaptcha.challenge",
			})
			return
		}

		// Stash the score in the context in case we need to publish it
		c.Set(CaptchaScore, res.Score)
	}
}

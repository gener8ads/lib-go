package request

import (
	"net"
	"net/http"
	"time"
)

// Client adds a some default timeouts to the standard http request lib
// More info: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
func Client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
		Timeout: time.Second * 30,
	}
}

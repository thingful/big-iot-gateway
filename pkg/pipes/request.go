package pipes

import (
	"fmt"

	"github.com/go-resty/resty"
)

// MakeRequest calls at url using token for Authentication
func MakeRequest(url string, token string) ([]byte, error) {
	client := resty.New()
	client.SetAuthToken(token)
	resp, err := client.R().Get(url)
	if err != nil || resp.StatusCode() != 200 {
		if err == nil {
			err = fmt.Errorf("status Code: %d received", resp.StatusCode())
		}
		return nil, err
	}
	return resp.Body(), nil
}

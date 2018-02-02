package pipes

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// MakeRequest calls at url using token for Authentication
func MakeRequest(url string, token string) ([]byte, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			err = errors.New(fmt.Sprintf("Status Code: %d received", resp.StatusCode))
		}
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

package utils

import (
	"io/ioutil"
	"net/http"
	"time"
)

const BaseURL = "https://console.cloudvbox.com/api2/json"
const AuthToken = "PVEAPIToken=root@pam!vbox=5639e614-5d8e-46e6-93ca-1a57a9238af0"

func FetchJSON(endpoint string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

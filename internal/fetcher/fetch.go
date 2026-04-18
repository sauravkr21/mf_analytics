package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type APIResponse struct {
	Meta struct {
		SchemeName string `json:"scheme_name"`
	} `json:"meta"`

	Data []struct {
		Date string `json:"date"`
		Nav  string `json:"nav"`
	} `json:"data"`
}

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func Fetch(code string) (APIResponse, error) {
	url := fmt.Sprintf("https://api.mfapi.in/mf/%s", code)

	resp, err := client.Get(url)
	if err != nil {
		return APIResponse{}, err
	}
	defer resp.Body.Close()

	// ✅ Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return APIResponse{}, fmt.Errorf("failed to fetch data: status %d", resp.StatusCode)
	}

	var result APIResponse

	// ✅ Check decode error
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return APIResponse{}, err
	}

	return result, nil
}
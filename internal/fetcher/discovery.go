package fetcher

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Scheme struct {
	SchemeCode string `json:"schemeCode"`
	SchemeName string `json:"schemeName"`
}

var allowedAMCs = []string{
	"icici", "hdfc", "axis", "sbi", "kotak",
}

func isValidAMC(name string) bool {
	for _, a := range allowedAMCs {
		if strings.Contains(name, a) {
			return true
		}
	}
	return false
}

func DiscoverSchemes() ([]Scheme, error) {
	resp, err := http.Get("https://api.mfapi.in/mf")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var all []Scheme
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil, err
	}

	var filtered []Scheme

	for _, s := range all {
		name := strings.ToLower(s.SchemeName)

		if isValidAMC(name) &&
			(strings.Contains(name, "mid cap") ||
				strings.Contains(name, "small cap")) &&
			strings.Contains(name, "direct") &&
			strings.Contains(name, "growth") {

			filtered = append(filtered, s)
		}
	}

	return filtered, nil
}
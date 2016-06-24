package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// FetchRelay will query the Onionoo API for a given Relay fingerprint and return a Relay object describing said relay
func FetchRelay(Fingerprint string) (Relay, error) {
	client := http.Client{}

	req, httpReqErr := http.NewRequest("GET", "https://onionoo.torproject.org/details?fingerprint="+Fingerprint, nil)

	if httpReqErr != nil {
		return Relay{Fingerprint: Fingerprint}, httpReqErr
	}

	req.Header.Set("User-Agent", "OnionWatch.email/1.0 (+https://onionwatch.email/bot/)")
	req.Header.Set("Content-Type", "application/json")

	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return Relay{Fingerprint: Fingerprint}, httpErr
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var ooResponse OnionooResponse
	jsonResponseParseErr := json.Unmarshal(body, &ooResponse)

	if jsonResponseParseErr != nil {
		return Relay{Fingerprint: Fingerprint}, jsonResponseParseErr
	}

	return ooResponse.Relays[0], nil

}

// FetchRelays works in a similar manner to FetchRelay but will instead return a slice of Relays relating to the given Fingerprints
// Note that we do not guarantee that the ordering of Relays will be the same as the order of Fingerprints from the slice
func FetchRelays(Fingerprints []string) ([]Relay, error) {
	client := http.Client{}
	var relays []Relay
	URLQueryString := ""
	FPCount := len(Fingerprints)
	for i := range Fingerprints {
		URLQueryString += Fingerprints[i]

		if i < FPCount {
			URLQueryString += ","
		}
	}

	req, httpReqErr := http.NewRequest("GET", "https://onionoo.torproject.org/details?fingerprint="+URLQueryString, nil)

	if httpReqErr != nil {
		return relays, httpReqErr
	}

	req.Header.Set("User-Agent", "OnionWatch.email/1.0 (+https://onionwatch.email/bot/)")
	req.Header.Set("Content-Type", "application/json")

	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return relays, httpErr
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var ooResponse OnionooResponse
	jsonResponseParseErr := json.Unmarshal(body, &ooResponse)

	if jsonResponseParseErr != nil {
		return relays, jsonResponseParseErr
	}

	return ooResponse.Relays, nil

}

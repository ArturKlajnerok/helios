package helios

import (
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	ServerAddress = "http://localhost:8080"
)

func skipInShortMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2e test in short mode.")
	}
}

func getResponse(address string, t *testing.T) []byte {
	resp, err := http.Get(address)
	if err != nil {
		t.Fatal("Error getting response", err)
	}

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		t.Fatal("Error reading response body", err)
	}
	resp.Body.Close()

	return body
}

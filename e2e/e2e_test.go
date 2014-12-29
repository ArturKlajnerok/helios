package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	ServerAddress = "http://localhost:8080"

	TokenEndpoint = "/token"

	TestUserName = "test"
	TestPassword = "test"

	TestClientId     = "123456"
	TestClientSecret = "aabbccdd"
)

func getJsonString(objmap map[string]*json.RawMessage, key string, t *testing.T) string {
	var str string
	v := objmap[key]
	if v == nil {
		return ""
	}
	if err := json.Unmarshal(*v, &str); err != nil {
		t.Fatal(fmt.Sprintf("Error unmarshalling value: %v", v), err)
	}

	return str
}

func unmarshall(jsonData []byte, t *testing.T) map[string]*json.RawMessage {
	var objmap map[string]*json.RawMessage
	err := json.Unmarshal(jsonData, &objmap)
	if err != nil {
		t.Fatal(fmt.Sprintf("Error unmarshalling data: %s", string(jsonData)), err)
	}
	return objmap
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

func getTokenResponseBody(userName string, password string, t *testing.T) []byte {
	address := ServerAddress + TokenEndpoint
	address = address + fmt.Sprintf("?grant_type=password&client_id=%s&client_secret=%s&username=%s&password=%s",
		TestClientId, TestClientSecret, userName, password)

	return getResponse(address, t)
}

func getTokenResponse(userName string, password string, t *testing.T) map[string]*json.RawMessage {

	body := getTokenResponseBody(userName, password, t)

	objmap := unmarshall(body, t)
	if getJsonString(objmap, "token_type", t) != "bearer" {
		t.Fatal(fmt.Sprintf("Invalid token data: %s", string(body)))
	}

	return objmap
}

func TestGetAccessToken(t *testing.T) {

	getTokenResponse(TestUserName, TestPassword, t)
}

func TestGetSameAccessToken(t *testing.T) {

	tokenResponse1 := getTokenResponse(TestUserName, TestPassword, t)
	tokenResponse2 := getTokenResponse(TestUserName, TestPassword, t)

	if getJsonString(tokenResponse1, "access_token", t) != getJsonString(tokenResponse2, "access_token", t) {
		t.Fatal(fmt.Sprintf("Received different tokens for same user: %s, %s",
			getJsonString(tokenResponse1, "access_token", t), getJsonString(tokenResponse2, "access_token", t)))
	}
}

func TestInvalidGetAccessToken(t *testing.T) {

	body := getTokenResponseBody(TestUserName, "InvalidPassword", t)

	objmap := unmarshall(body, t)
	if getJsonString(objmap, "error", t) != "access_denied" {
		t.Fatal(fmt.Sprintf("Received invalid access denied answer: %s", string(body)))
	}
}

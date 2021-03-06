package helios

import (
	"encoding/json"
	"fmt"
	"testing"
)

const (
	TokenEndpoint = "/token"

	TestUserName = "test"
	TestPassword = "test"

	TestClientId     = "123456"
	TestClientSecret = "aabbccdd"
)

func getJsonString(objMap map[string]*json.RawMessage, key string, t *testing.T) string {
	var str string
	v := objMap[key]
	if v == nil {
		return ""
	}
	if err := json.Unmarshal(*v, &str); err != nil {
		t.Fatal(fmt.Sprintf("Error unmarshalling value: %v", v), err)
	}

	return str
}

func unmarshall(jsonData []byte, t *testing.T) map[string]*json.RawMessage {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(jsonData, &objMap)
	if err != nil {
		t.Fatal(fmt.Sprintf("Error unmarshalling data: %s", string(jsonData)), err)
	}
	return objMap
}

func getTokenResponseBody(userName string, password string, t *testing.T) []byte {
	address := ServerAddress + TokenEndpoint
	address = address + fmt.Sprintf("?grant_type=password&client_id=%s&client_secret=%s&username=%s&password=%s",
		TestClientId, TestClientSecret, userName, password)

	return getResponse(address, t)
}

func getTokenResponse(userName string, password string, t *testing.T) map[string]*json.RawMessage {

	body := getTokenResponseBody(userName, password, t)

	objMap := unmarshall(body, t)
	if getJsonString(objMap, "token_type", t) != "bearer" {
		t.Fatal(fmt.Sprintf("Invalid token data: %s", string(body)))
	}

	return objMap
}

func TestE2eGetAccessToken(t *testing.T) {
	skipInShortMode(t)

	getTokenResponse(TestUserName, TestPassword, t)
}

func TestE2eGetSameAccessToken(t *testing.T) {
	skipInShortMode(t)

	tokenResponse1 := getTokenResponse(TestUserName, TestPassword, t)
	tokenResponse2 := getTokenResponse(TestUserName, TestPassword, t)

	if getJsonString(tokenResponse1, "access_token", t) != getJsonString(tokenResponse2, "access_token", t) {
		t.Fatal(fmt.Sprintf("Received different tokens for same user: %s, %s",
			getJsonString(tokenResponse1, "access_token", t), getJsonString(tokenResponse2, "access_token", t)))
	}
}

func TestE2eInvalidGetAccessToken(t *testing.T) {
	skipInShortMode(t)

	body := getTokenResponseBody(TestUserName, "InvalidPassword", t)

	objMap := unmarshall(body, t)
	if getJsonString(objMap, "error", t) != "access_denied" {
		t.Fatal(fmt.Sprintf("Received invalid access denied answer: %s", string(body)))
	}
}

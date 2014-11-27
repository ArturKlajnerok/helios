package models

import (
	"testing"
)

const (
	USER_ID            = 1
	USER_PASSWORD      = "test"
	USER_PASSWORD_HASH = "c4d81016667031737ffeda045105816e"
)

func TestValidPassword(t *testing.T) {
	user := User{Id: USER_ID, HashedPassword: USER_PASSWORD_HASH}
	if !user.ValidPassword(USER_PASSWORD) {
		t.Error("Expected hash: ", USER_PASSWORD_HASH)
	}
}

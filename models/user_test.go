package models

import (
	"testing"
)

const (
	UserID           = 1
	UserPassword     = "test"
	UserPasswordHash = "c4d81016667031737ffeda045105816e"
)

func TestIsValidPassword(t *testing.T) {
	user := User{Id: UserID, HashedPassword: UserPasswordHash}
	if !user.IsValidPassword(UserPassword) {
		t.Error("Expected hash: ", UserPasswordHash)
	}
}

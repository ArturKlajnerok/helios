package models

import (
	"testing"
)

const (
	UserID              = 1
	UserPassword        = "test"
	UserSalt            = "abcd1234"
	UserPasswordHash    = "d388ff04ec4afef192699e5b44a1b7ea"
	UserOldPasswordHash = "c4d81016667031737ffeda045105816e"
)

func TestIsValidPassword(t *testing.T) {
	saltedHash := HashPrefix + UserSalt + ":" + UserPasswordHash
	user := User{Id: UserID, HashedPassword: saltedHash}
	if !user.IsValidPassword(UserPassword) {
		t.Error("Wrong passoword hash. Expected: ", saltedHash)
	}
	userOld := User{Id: UserID, HashedPassword: UserOldPasswordHash}
	if !userOld.IsValidPassword(UserPassword) {
		t.Error("Wrong old passoword hash. Expected: ", UserOldPasswordHash)
	}
}

func TestHashPassword(t *testing.T) {
	hash := HashPassword(UserPassword, UserSalt)
	if hash != UserPasswordHash {
		t.Error("Wrong hash generated. Expected:", UserPasswordHash, "Actual:", hash)
	}
}

func TestOldHashPassword(t *testing.T) {
	hash := OldHashPassword(UserPassword, UserID)
	if hash != UserOldPasswordHash {
		t.Error("Wrong hash generated. Expected:", UserOldPasswordHash, "Actual:", hash)
	}
}

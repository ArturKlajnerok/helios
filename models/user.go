package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/coopernurse/gorp"
	"time"
)

type User struct {
	Id                 int64      `db:"user_id"`
	Name               string     `db:"user_name"`
	RealName           string     `db:"user_real_name"`
	Password           string     `db:"-"`
	HashedPassword     string     `db:"user_password"`
	NewPassword        string     `db:"user_newpassword"`
	NewPassTime        *string    `db:"user_newpass_time"`
	Email              string     `db:"user_email"`
	Touched            string     `db:"user_touched"`
	Token              string     `db:"user_token"`
	EmailAuthenticated *string    `db:"user_email_authenticated"`
	EmailToken         *string    `db:"user_email_token"`
	EmailTokenExpires  *string    `db:"user_email_token_expires"`
	Registration       *string    `db:"user_registration"`
	EditCount          *int64     `db:"user_editcount"`
	BirthDate          *time.Time `db:"user_birthdate"`
	Options            []byte     `db:"user_options"`
}

func (user *User) IsValidPassword(password string) bool {
	hash := hashPassword(password, user.Id)
	return user.HashedPassword == hash
}

func hashPassword(password string, userId int64) string {
	hasher := md5.New()
	hasher.Write([]byte(password))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hash = fmt.Sprintf("%d-%s", userId, hash)

	hasher.Reset()
	hasher.Write([]byte(hash))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (user *User) FindByName(dbmap *gorp.DbMap) bool {
	err := dbmap.SelectOne(&user, "select * from user where user_name=?", user.Name)
	if err != nil {
		return false
	}
	return true
}

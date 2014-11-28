package models

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/coopernurse/gorp"
	"strconv"
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

func (user *User) ValidPassword(password string) bool {
	hasher := md5.New()
	hasher.Write([]byte(password))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hash = strconv.FormatInt(user.Id, 10) + "-" + hash

	hasher.Reset()
	hasher.Write([]byte(hash))
	hash = hex.EncodeToString(hasher.Sum(nil))
	return user.HashedPassword == hash
}

func (user *User) FindByName(dbmap *gorp.DbMap) bool {
	err := dbmap.SelectOne(&user, "select * from user where user_name=?", user.Name)
	if err != nil {
		return false
	}
	return true
}

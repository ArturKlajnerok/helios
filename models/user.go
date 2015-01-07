package models

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/Wikia/go-commons/logger"
	"github.com/coopernurse/gorp"
	"log"
	"strings"
	"time"
)

const (
	HashPrefix = ":B:"
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
	if len(user.HashedPassword) < 3 {
		log.Printf("To short hash for user: %s\n", user.Name)
		return false
	}

	prefix := user.HashedPassword[0:3]
	if prefix == HashPrefix {
		salt, passHash := ExtractHashAndSalt(user.HashedPassword)
		return passHash == HashPassword(password, salt)
	}

	return user.HashedPassword == OldHashPassword(password, user.Id)
}

func ExtractHashAndSalt(hash string) (string, string) {
	splitedHash := strings.SplitN(hash[3:], ":", 2)
	if len(splitedHash) != 2 {
		log.Printf("Can't split properly hash: %s\n", hash)
		return "", ""
	}
	return splitedHash[0], splitedHash[1]
}

func HashPassword(password string, salt string) string {
	hasher := md5.New()
	hasher.Write([]byte(password))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hash = fmt.Sprintf("%s-%s", salt, hash)

	hasher.Reset()
	hasher.Write([]byte(hash))
	return hex.EncodeToString(hasher.Sum(nil))
}

func OldHashPassword(password string, userId int64) string {
	hasher := md5.New()
	hasher.Write([]byte(password))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hash = fmt.Sprintf("%d-%s", userId, hash)

	hasher.Reset()
	hasher.Write([]byte(hash))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (user *User) FindByName(dbmap *gorp.DbMap) (bool, error) {
	err := dbmap.SelectOne(&user, "select * from user where user_name=?", user.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		} else {
			logger.GetLogger().ErrorErr(err)
			return false, err
		}
	}
	return true, nil
}

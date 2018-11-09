package server

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// Token struct
type Token struct {
	id string
	db TokenSQL
}

// TokenSQL struct
type TokenSQL struct {
	UID        int    `db:"user_id"`
	ID         int    `db:"token_id"`
	Type       string `db:"token_type"`
	State      int    `db:"token_state"`
	Name       string `db:"token_name"`
	TimeStart  int    `db:"token_time_start"`
	TimeExpire int    `db:"token_time_expire"`
}

func (t *Token) generate() (string, error) {
	i := fmt.Sprintf("%d", time.Now().Unix())

	res, err := t.baseConvert(i, 30, 33)
	if err != nil {
		return "", err
	}

	t.id = t.uniqid(res)

	return t.id, nil
}

func (t *Token) new(tokenType string, userID int) (string, error) {
	if _, err := t.generate(); err != nil {
		return "", err
	}

	if err := t.revokeOfType(tokenType, userID, userID); err != nil {
		return "", err
	}

	if _, err = conn.Exec("insert into `tokens` (`user_id`, `token_state`, `token_type`, `token_name`, `token_time_start`) values (?, '1', ?, ?, ?)",
		userID,
		tokenType,
		t.id,
		time.Now().Unix()); err != nil {
		return "", err
	}

	return t.id, nil
}

func (t *Token) revokeOfType(tokenType string, userID int, revokeUserID int) error {
	if _, err = conn.Exec("update `tokens` set `token_state` = '0', `token_time_expire` = ?, `revoke_id` = ? where `token_state` = '1' and `token_type` = ? and `user_id` = ?", time.Now().Unix(), revokeUserID, tokenType, userID); err != nil {
		return err
	}

	return nil
}

func (t *Token) revoke(tokenName string, revokeUserID int) error {
	t.validate(tokenName)

	if _, err = conn.Exec("update `tokens` set `token_state` = '0', `token_time_expire` = ?, `revoke_id` = ? where `token_id` = ?", time.Now().Unix(), revokeUserID, t.db.ID); err != nil {
		return err
	}

	return nil
}

func (t *Token) validate(tokenName string) error {
	if t.db.ID == 0 {
		if err := t.find(tokenName, t.db.Type); err != nil {
			return err
		}
	}

	if t.db.State == 0 {
		return fmt.Errorf("Token has been revoked")
	}

	return nil
}

func (t *Token) find(tokenName string, tokenType string) error {
	if err := conn.Unsafe().Get(&t.db, "select * from `tokens` where `token_name` = ? and (`token_type` = ? or '' = ?)", tokenName, tokenType, tokenType); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("Token not found")
		}

		return err
	}

	return nil
}

func (t *Token) baseConvert(number string, frombase, tobase int) (string, error) {
	i, err := strconv.ParseInt(number, frombase, 0)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(i, tobase), nil
}

func (t *Token) uniqid(prefix string) string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000
	return fmt.Sprintf("%s%08x%05x", prefix, sec, usec)
}

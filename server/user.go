package server

import (
	"database/sql"
)

type userDB struct {
}

type user struct {
	ID          string `db:"user_id"`
	Token       string `db:"user_token"`
	Repo        string `db:"user_repo_name"`
	RepoTime    int    `db:"user_repo_time"`
	RepoChannel string `db:"user_repo_channel"`
}

func (u *user) find(id string) error {
	if err := conn.Get(u, "select * from users where user_id = $1", id); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

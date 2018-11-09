package server

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/dhalturin/release-manager/data"
	"github.com/dhalturin/release-manager/slack"
)

func repoList(isAdmin bool) ([]repo, error) {
	Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
		Attachments: []slack.ResponseAttachment{
			slack.ResponseAttachment{
				Color: "#bff442",
				Text:  "Loading your repositories :nyancat:",
			},
		},
	})

	sql := ""
	if isAdmin {
		sql = fmt.Sprintf("select * from repos where repo_admin = '%s' and repo_channel = '%s'", Payload.User.ID, Payload.Channel.ID)
	} else {
		sql = fmt.Sprintf("select * from repos where repo_channel = '%s'", Payload.Channel.ID)
	}

	repos := []repo{}
	if err := conn.Select(&repos, sql); err != nil {
		return nil, err
	}

	return repos, nil
}

func repoListView(isAdmin bool, callback string, withConfirm bool) error {
	repos, err := repoList(isAdmin)
	if err != nil {
		return err
	}

	options := &[]slack.ResponseAttachmentActionsOptions{}
	for _, repo := range repos {
		if isAdmin {
			if Payload.User.ID != repo.Admin {
				continue
			}
		} else {
			if res, err := Slack.UserGroupFindUser(repo.Permission, Payload.User.ID); err != nil {
				log.Panicln(">>>> ERROR: ", err.Error())
				continue
			} else if !res {
				continue
			}
		}

		*options = append(*options, slack.ResponseAttachmentActionsOptions{
			Text:  repo.Name,
			Value: repo.Name,
		})
	}

	if len(*options) == 0 {
		Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
			Attachments: []slack.ResponseAttachment{
				slack.ResponseAttachment{
					Color: "#4fb1e2",
					Text:  "You do not have any repository for this channel",
				},
			},
		})

		return nil
	}

	actions := []slack.ResponseAttachmentActions{}

	if withConfirm {
		actions = append(actions, slack.ResponseAttachmentActions{
			Name:    "repo-list",
			Type:    "select",
			Style:   "primary",
			Options: options,
			Confirm: &slack.ResponseConfirm{
				TextDismiss: "No",
				TextOK:      "Yep",
			},
		})
	} else {
		actions = append(actions, slack.ResponseAttachmentActions{
			Name:    "repo-list",
			Type:    "select",
			Style:   "primary",
			Options: options,
		})
	}

	Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
		Attachments: []slack.ResponseAttachment{
			slack.ResponseAttachment{
				Callback: callback,
				Color:    "#4fb1e2",
				Text:     "Choose repository:",
				Actions:  actions,
			},
		},
	})

	return nil
}

func repoFind(repoName string) (repo, error) {
	Error := Err{Pointer: "repoFind", Meta: "func"}

	repo := repo{}
	if err := conn.Get(&repo, "select * from repos where repo_name = $1 ", repoName); err != nil {
		if err != sql.ErrNoRows {
			return repo, err
		}
	}

	if repo.ID == 0 {
		return repo, fmt.Errorf(Error.find(1120).Detail)
	}

	return repo, nil
}

func repoUsed(userID, channelID string) (string, error) {
	user := user{}
	if err := conn.Get(&user, "select * from users where user_id = $1 and extract(epoch from now()) - user_repo_time < $2 and user_repo_channel = $3", userID, (data.RepoTTL * 60), channelID); err != nil {
		if err != sql.ErrNoRows {
			return user.Repo, err
		}
	}

	return user.Repo, nil
}

func repoUsedUpdate(userID, channelID, repoName string) error {
	if _, err = conn.Exec("update users set user_repo_channel = $2, user_repo_name = $3, user_repo_time = round(extract(epoch from now())) where user_id = $1", userID, channelID, repoName); err != nil {
		return err
	}

	return nil
}

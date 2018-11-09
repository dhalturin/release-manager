package server

import (
	"fmt"

	"github.com/dhalturin/release-manager/slack"
	gitlab "github.com/xanzy/go-gitlab"
)

func init() {
	method := "init"

	MethodList[method] = MethodStruct{
		NoToken:   true,
		OnlyGuest: true,
		Help:      []string{method, "initialize this tool for current user"},
		Func: func() (interface{}, bool) {
			Error := Err{Pointer: method, Meta: "method"}

			if Payload.UserToken != "" {
				return "> Your personal access token is already was added :tada:", false
			}

			if Payload.Callback == "init value" {
				git := gitlab.NewClient(nil, Payload.Submission.Token)
				git.SetBaseURL(GitLabHost)
				if _, _, err := git.Version.GetVersion(); err != nil {
					return Error.get([]int{1111}).join(), true
				}

				user := user{}
				if err := user.find(Payload.User.ID); err != nil {
					return Error.push(500, 100, err.Error()).join(), true
				}

				query := ""
				if user.ID == "" {
					query = fmt.Sprintf("insert into users (user_id, user_token) values ('%s', '%s')", Payload.User.ID, Payload.Submission.Token)
				} else {
					query = fmt.Sprintf("update users set user_token = '%s' where user_id = '%s'", Payload.User.ID, Payload.Submission.Token)
				}

				if _, err := conn.Exec(query); err != nil {
					return err.Error(), true
				}

				Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
					Attachments: []slack.ResponseAttachment{
						slack.ResponseAttachment{
							Color: "good",
							Text:  "Your GitLab access token is successfully added",
						},
					},
				})

				return "", false
			}

			// Default action

			if err := Slack.DialogOpen(Payload.TriggerID, slack.Dialog{
				CallbackID: "init value",
				Title:      "Initialisation",
				State:      "form-submit",
				Elements: []slack.Elements{
					slack.Elements{
						Type:  "text",
						Label: "Personal access token",
						Name:  "token",
						Hint:  fmt.Sprintf("Your personal GitLab access token from:\n%s/profile/personal_access_tokens", GitLabHost),
					},
				},
			}); err != nil {
				return Error.push(500, 101, err.Error()).join(), true
			}

			return "", false
		},
	}
}

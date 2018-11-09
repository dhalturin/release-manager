package server

import (
	"database/sql"
	"fmt"

	"github.com/dhalturin/release-manager/slack"
	gitlab "github.com/xanzy/go-gitlab"
)

func init() {
	method := "repo add"

	MethodList[method] = MethodStruct{
		Help: []string{method, "adding new repository"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				if Payload.Callback == "repo add value" {
					repo := repo{}
					if err := conn.Get(&repo, "select * from repos where repo_name = $1", Payload.Submission.Repository); err != nil {
						if err != sql.ErrNoRows {
							SendError(Error.push(500, 100, err.Error()).join())
							return
						}
					}

					if repo.ID > 0 {
						SendError(Error.get([]int{1110}).join())
						return
					}

					git := gitlab.NewClient(nil, Payload.UserToken)
					git.SetBaseURL(GitLabHost)

					if _, _, err := git.Projects.GetProject(Payload.Submission.Repository); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if _, err := conn.Exec("insert into repos (repo_name, repo_ref, repo_prefix, repo_channel, repo_permission, repo_admin) values ($1, $2, $3, $4, $5, $6)",
						Payload.Submission.Repository,
						Payload.Submission.Ref,
						Payload.Submission.Prefix,
						Payload.Submission.Channel,
						Payload.Submission.Permission,
						Payload.User.ID,
					); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
						Attachments: []slack.ResponseAttachment{
							slack.ResponseAttachment{
								Color: "good",
								Text:  "The repository is successfully added",
							},
						},
					})

					return
				}

				// Default action

				usergroups, err := Slack.UserGroupList()
				if err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				}

				Slack.Response(Payload.ResponseURL, slack.ResponseJSON{Text: "", Attachments: []slack.ResponseAttachment{
					slack.ResponseAttachment{
						Color: "#4fb1e2",
						Text:  "Fill in the form to add a repository",
					},
				}})

				options := []slack.Options{}
				for _, group := range usergroups {
					options = append(options, slack.Options{
						Label: fmt.Sprintf("%s (members: %d)", group.Handle, group.UserCount),
						Value: group.ID,
					})
				}

				elements := []slack.Elements{}

				elements = append(elements, slack.Elements{
					Type:       "select",
					Label:      "Channel",
					Name:       "channel",
					Value:      Payload.Channel.ID,
					Hint:       "Channel for notify about new release",
					DataSource: "channels",
				})

				elements = append(elements, slack.Elements{
					Type:        "text",
					Label:       "GitLab repository",
					Name:        "repository",
					Placeholder: "namespace/repository",
					Hint:        "Repository with namespace",
				})

				elements = append(elements, slack.Elements{
					Type:        "text",
					Label:       "Repository ref",
					Value:       "master",
					Name:        "ref",
					Placeholder: "master",
					Hint:        "The branch name to create release branch from",
				})

				elements = append(elements, slack.Elements{
					Type:    "select",
					Label:   "Permission",
					Name:    "permission",
					Hint:    "Select which user group will be able to use *Release Manager*",
					Options: options,
				})

				elements = append(elements, slack.Elements{
					Type:        "text",
					Label:       "Prefix",
					Name:        "prefix",
					Placeholder: "",
					Optional:    true,
					Hint:        "The prefix of the message header of the release. Optional",
				})

				if err := Slack.DialogOpen(Payload.TriggerID, slack.Dialog{
					CallbackID: "repo add value",
					Title:      "Adding repository",
					State:      "form-submit",
					Elements:   elements,
				}); err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				}
			}()

			return "", false
		},
	}
}

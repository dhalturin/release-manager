package server

import (
	"fmt"

	"github.com/dhalturin/release-manager/slack"
)

func init() {
	method := "repo edit"

	MethodList[method] = MethodStruct{
		Help: []string{method, "update existing repository"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				if err := repoListView(true, "repo edit value", false); err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				}

				if Payload.Callback == "repo edit value" {
					repo, err := repoFind(Payload.Actions[0].Selected[0].Value)
					if err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

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
						Value:      repo.Channel,
						Hint:       "Channel for notify about new release",
						DataSource: "channels",
					})

					// elements = append(elements, slack.Elements{
					// 	Type:        "text",
					// 	Label:       "GitLab repository",
					// 	Name:        "repository",
					// 	Placeholder: "namespace/repository",
					// 	Hint:        "Repository with namespace",
					// 	Value:       repo.Name,
					// })

					elements = append(elements, slack.Elements{
						Type:        "text",
						Label:       "Repository ref",
						Value:       repo.Ref,
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
						Value:   repo.Permission,
					})

					elements = append(elements, slack.Elements{
						Type:        "text",
						Label:       "Prefix",
						Name:        "prefix",
						Placeholder: "",
						Optional:    true,
						Hint:        "The prefix of the message header of the release. Optional",
						Value:       repo.Prefix,
					})

					if err := Slack.DialogOpen(Payload.TriggerID, slack.Dialog{
						CallbackID: "repo edit value-repo",
						Title:      "Adding repository",
						State:      repo.ID,
						Elements:   elements,
					}); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					return
				} else if Payload.Callback == "repo edit value-repo" {
					if _, err := conn.Exec("update repos set repo_channel = $1, repo_ref = $2, repo_permission = $3, repo_prefix = $4 where repo_id = $5",
						Payload.Submission.Channel,
						Payload.Submission.Ref,
						Payload.Submission.Permission,
						Payload.Submission.Prefix,
						Payload.State,
					); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
						Attachments: []slack.ResponseAttachment{
							slack.ResponseAttachment{
								Color: "good",
								Text:  "The repository is successfully updated",
							},
						},
					})

					return
				}

				// Default action

				if err := repoListView(true, "repo edit value", false); err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				}
			}()

			return "", false
		},
	}
}

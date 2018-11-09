package server

import (
	"encoding/json"
	"fmt"

	"github.com/dhalturin/release-manager/slack"
)

func init() {
	method := "release publish"

	MethodList[method] = MethodStruct{
		Help: []string{method, "publish release information to channel"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				repoName, err := repoUsed(Payload.User.ID, Payload.Channel.ID)
				if err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				} else if repoName == "" {
					if Payload.Callback == "release publish repo-list" {
						repoName = Payload.Actions[0].Selected[0].Value
					} else {
						if err := repoListView(false, "release publish repo-list", false); err != nil {
							SendError(Error.push(500, 100, err.Error()).join())
							return
						}
					}
				} else {
					Slack.ResponseDelay(slack.ResponseJSON{
						User:    Payload.User.ID,
						Channel: Payload.Channel.ID,
						Attachments: []slack.ResponseAttachment{
							slack.ResponseAttachment{
								Color: "good",
								Text:  fmt.Sprintf("Using repository *%s*", repoName),
							},
						},
					})
				}

				if repoName != "" {
					repo, err := repoFind(repoName)
					if err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if err := repoUsedUpdate(Payload.User.ID, Payload.Channel.ID, repo.Name); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if repo.Release == 0 {
						Slack.Response(Payload.ResponseURL, slack.ResponseJSON{Text: "> No release has been created for the current repository"})
						return
					}

					if repo.Tasks == "" {
						SendError(Error.get([]int{1131}).join())
						return
					}

					if repo.Prefix == "" {
						repo.Prefix = repo.Name
					}

					tasks := []string{}
					if err := json.Unmarshal([]byte(repo.Tasks), &tasks); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					fields := []slack.ResponseAttachmentField{}
					for _, j := range tasks {
						fields = append(fields, slack.ResponseAttachmentField{
							Value: j,
						})
					}

					attachments := []slack.ResponseAttachment{
						slack.ResponseAttachment{
							Text:   fmt.Sprintf("*%s* `release-%d` includes:", repo.Prefix, repo.Release),
							Footer: fmt.Sprintf("Publisher: <@%s>", Payload.User.ID),
							Fields: fields,
						},
					}

					if len(Payload.Actions) == 0 || Payload.Actions[0].Name != "publish" {
						attachments = append(attachments, slack.ResponseAttachment{
							Callback: "release publish repo",
							Actions: []slack.ResponseAttachmentActions{
								slack.ResponseAttachmentActions{
									Name:  "publish",
									Value: repoName,
									Text:  "publish",
									Type:  "button",
									Style: "primary",
								},
							},
						})
					}

					if len(Payload.Actions) > 0 && Payload.Actions[0].Name == "publish" {
						Slack.ChatPost(slack.ChatPostMessageJSON{
							Channel:     Payload.Channel.ID,
							Attachments: attachments,
						})
					} else {
						Slack.ResponseDelay(slack.ResponseJSON{
							User:        Payload.User.ID,
							Channel:     Payload.Channel.ID,
							Attachments: attachments,
						})
					}

					return
				}
			}()

			return "", false
		},
	}
}

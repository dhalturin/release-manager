package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dhalturin/release-manager/slack"
)

func init() {
	method := "release edit"

	MethodList[method] = MethodStruct{
		Help: []string{method, "edit information of current release (only notes)"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				repoName, err := repoUsed(Payload.User.ID, Payload.Channel.ID)
				if err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				} else if repoName == "" {
					if Payload.Callback == "release edit repo-list" {
						repoName = Payload.Actions[0].Selected[0].Value
					} else {
						if err := repoListView(false, "release edit repo-list", false); err != nil {
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

				if Payload.Callback == "release edit repo-task" {
					tasks := []byte{}

					if len(Payload.Submission.Tasks) > 0 {
						res, err := json.Marshal(strings.Split(Payload.Submission.Tasks, "\n"))
						if err != nil {
							SendError(Error.push(500, 100, err.Error()).join())
							return
						}

						tasks = res
					}

					if _, err := conn.Exec("update repos set repo_tasks = $1 where repo_id = $2", string(tasks), Payload.State); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.ResponseDelay(slack.ResponseJSON{
						User:    Payload.User.ID,
						Channel: Payload.Channel.ID,
						Attachments: []slack.ResponseAttachment{
							slack.ResponseAttachment{
								Color: "good",
								Text:  "Release successfully updated",
							},
						},
					})

					return
				} else if repoName != "" {
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

					tasks := []string{}

					if len(repo.Tasks) > 0 {
						if err := json.Unmarshal([]byte(repo.Tasks), &tasks); err != nil {
							SendError(Error.push(500, 100, err.Error()).join())
							return
						}
					}

					if err := Slack.DialogOpen(Payload.TriggerID, slack.Dialog{
						CallbackID: "release edit repo-task",
						Title:      fmt.Sprintf("Edit release-%d", repo.Release),
						State:      repo.ID,
						Elements: []slack.Elements{
							slack.Elements{
								Type:     "textarea",
								Label:    "Task list",
								Value:    strings.Join(tasks, "\n"),
								Name:     "tasks",
								Hint:     "A list of task. One task per lines.",
								Optional: true,
							},
						},
					}); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					return
				}
			}()

			return "", false
		},
	}
}

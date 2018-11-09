package server

import (
	"fmt"

	"github.com/dhalturin/release-manager/slack"
	gitlab "github.com/xanzy/go-gitlab"
)

func init() {
	method := "release del"

	MethodList[method] = MethodStruct{
		Help: []string{method, "removing release branch"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				repoName, err := repoUsed(Payload.User.ID, Payload.Channel.ID)
				if err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				} else if repoName == "" {
					if Payload.Callback == "release del repo-list" {
						repoName = Payload.Actions[0].Selected[0].Value
					} else {
						if err := repoListView(false, "release del repo-list", false); err != nil {
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

					git := gitlab.NewClient(nil, Payload.UserToken)
					git.SetBaseURL(GitLabHost)

					project, _, err := git.Projects.GetProject(repo.Name)
					if err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					release := fmt.Sprintf("release-%d", repo.Release)

					if _, err := git.Branches.DeleteBranch(project.ID, release); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if _, err := conn.Exec("update repos set repo_release = '0', repo_tasks = '' where repo_id = $1", repo.ID); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.ResponseDelay(slack.ResponseJSON{
						User:    Payload.User.ID,
						Channel: Payload.Channel.ID,
						Text:    fmt.Sprintf("> Release \"*%s*\" is successfully removed", release),
					})

					return
				}
			}()

			return "", false
		},
	}
}

package server

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/dhalturin/release-manager/slack"
	gitlab "github.com/xanzy/go-gitlab"
)

func init() {
	method := "release add"

	MethodList[method] = MethodStruct{
		Help: []string{method, "create new release branch"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				repoName, err := repoUsed(Payload.User.ID, Payload.Channel.ID)
				if err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				} else if repoName == "" {
					if Payload.Callback == "release add repo-list" {
						repoName = Payload.Actions[0].Selected[0].Value
					} else {
						if err := repoListView(false, "release add repo-list", false); err != nil {
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

					git := gitlab.NewClient(nil, Payload.UserToken)
					git.SetBaseURL(GitLabHost)

					project, _, err := git.Projects.GetProject(repo.Name)
					if err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if repo.Release == 0 {
						Slack.ResponseDelay(slack.ResponseJSON{
							User:    Payload.User.ID,
							Channel: Payload.Channel.ID,
							Attachments: []slack.ResponseAttachment{
								slack.ResponseAttachment{
									Color: "#bff442",
									Text:  "Loading list of branches :nyancat:",
								},
							},
						})

						branches := []BranchList{}
						if err := gitLabListBranches(git, project.ID, &branches, "release-([0-9]+)"); err != nil {
							SendError(Error.push(500, 100, err.Error()).join())
							return
						}

						if len(branches) > 0 {
							num, err := strconv.Atoi(regexp.MustCompile("release-").ReplaceAllString(branches[0].Name, ""))
							if err != nil {
								SendError(Error.push(500, 100, err.Error()).join())
								return
							}

							repo.Release = num
						}
					}

					repo.Release++

					branch := fmt.Sprintf("release-%d", repo.Release)
					if _, _, err := git.Branches.CreateBranch(repo.Name, &gitlab.CreateBranchOptions{
						Branch: &branch,
						Ref:    &repo.Ref,
					}); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if _, err := conn.Exec("update repos set repo_release = $1, repo_tasks = '' where repo_id = $2", repo.Release, repo.ID); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.ResponseDelay(slack.ResponseJSON{
						User:    Payload.User.ID,
						Channel: Payload.Channel.ID,
						Attachments: []slack.ResponseAttachment{
							slack.ResponseAttachment{
								Color: "#4fb1e2",
								Text:  fmt.Sprintf("Branch for release was created: *%d*", repo.Release),
							},
						},
					})

					return
				}
			}()

			return "", false
		},
	}
}

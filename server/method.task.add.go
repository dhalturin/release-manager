package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dhalturin/release-manager/slack"
	gitlab "github.com/xanzy/go-gitlab"
)

func init() {
	method := "task add"

	MethodList[method] = MethodStruct{
		Help: []string{method, "add task branch to release branch"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				repoName, err := repoUsed(Payload.User.ID, Payload.Channel.ID)
				if err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				} else if repoName == "" {
					if Payload.Callback == "task add repo-list" {
						repoName = Payload.Actions[0].Selected[0].Value
					} else {
						if err := repoListView(false, "task add repo-list", false); err != nil {
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

				if Payload.Callback == "task add task-list" {
					repo, err := repoFind(Payload.State)
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

					tasksMap := map[string]string{}
					tasks := []string{}

					if len(repo.Tasks) > 0 {
						if err := json.Unmarshal([]byte(repo.Tasks), &tasks); err != nil {
							SendError(Error.push(500, 100, err.Error()).join())
							return
						}
					}

					for _, task := range strings.Split(Payload.Submission.Tasks, ",") {
						tasksMap[strings.Trim(task, " ")] = ""
					}

					release := fmt.Sprintf("release-%d", repo.Release)

					attachments := []slack.ResponseAttachment{}
					for task := range tasksMap {
						Slack.ResponseDelay(slack.ResponseJSON{
							Channel: Payload.Channel.ID,
							User:    Payload.User.ID,
							Attachments: []slack.ResponseAttachment{
								slack.ResponseAttachment{
									Color: "#bff442",
									Text:  fmt.Sprintf("Processing task: *%s*", task),
								},
							},
						})

						branch, _, err := git.Branches.GetBranch(project.ID, task)
						if err != nil {
							tasksMap[task] = err.Error()
						} else {
							opts := gitlab.CreateMergeRequestOptions{
								Title:        &branch.Commit.Title,
								SourceBranch: &branch.Name,
								TargetBranch: &release,
							}

							if mr, _, err := git.MergeRequests.CreateMergeRequest(project.ID, &opts); err == nil {
								if _, _, err := git.MergeRequests.AcceptMergeRequest(project.ID, mr.IID, &gitlab.AcceptMergeRequestOptions{}); err != nil {
									tasksMap[task] = err.Error()
								}
							} else {
								tasksMap[task] = err.Error()
							}
						}

						if tasksMap[task] == "" {
							tasksMap[task] = "merged"

							tasks = append(tasks, branch.Commit.Title)
						}

						attachments = append(attachments, slack.ResponseAttachment{
							Color: "#4fb1e2",
							Text:  fmt.Sprintf("%s: %s", task, tasksMap[task]),
						})
					}

					Slack.ResponseDelay(slack.ResponseJSON{
						Channel: Payload.Channel.ID,
						User:    Payload.User.ID,
						Text:    "> Finding and aborting pipelines for current release",
					})

					if pipelines, _, err := git.Pipelines.ListProjectPipelines(project.ID, &gitlab.ListProjectPipelinesOptions{Ref: &release}); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
					} else {
						for _, pipeline := range pipelines {
							git.Pipelines.CancelPipelineBuild(project.ID, pipeline.ID)
						}
					}

					if len(attachments) > 0 {
						tasks, err := json.Marshal(tasks)
						if err != nil {
							SendError(Error.push(500, 100, fmt.Sprintf("marshal task is failure: %s; %+v", err.Error(), string(tasks))).join())
						}

						if _, err := conn.Exec("update repos set repo_pipeline = '0', repo_tasks = $1 where repo_id = $2", tasks, repo.ID); err != nil {
							SendError(Error.push(500, 100, fmt.Sprintf("update task is failure: %s; %+v", err.Error(), string(tasks))).join())
						}

						Slack.ResponseDelay(slack.ResponseJSON{
							Channel:     Payload.Channel.ID,
							User:        Payload.User.ID,
							Text:        fmt.Sprintf("> Merge tasks to *release-%d*:", repo.Release),
							Attachments: attachments,
						})
					}

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

					git := gitlab.NewClient(nil, Payload.UserToken)
					git.SetBaseURL(GitLabHost)

					if _, _, err := git.Projects.GetProject(repo.Name); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.ResponseDelay(slack.ResponseJSON{
						User:    Payload.User.ID,
						Channel: Payload.Channel.ID,
						Attachments: []slack.ResponseAttachment{
							slack.ResponseAttachment{
								Color: "#4fb1e2",
								Text:  "Fill in the form to add a tasks",
							},
						},
					})

					if err := Slack.DialogOpen(Payload.TriggerID, slack.Dialog{
						CallbackID: "task add task-list",
						Title:      fmt.Sprintf("Release-%d", repo.Release),
						State:      repo.Name,
						Elements: []slack.Elements{
							slack.Elements{
								Type:        "text",
								Label:       "Tasks",
								Name:        "tasks",
								Placeholder: "task-1, task-2, task-3",
								Hint:        "list of branches with tasks through the a comma",
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

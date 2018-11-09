package server

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/dhalturin/release-manager/slack"
	gitlab "github.com/xanzy/go-gitlab"
)

func init() {
	method := "job abort"

	MethodList[method] = MethodStruct{
		Help: []string{method, "abort running jobs from last pipeline of current release"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				repoName, err := repoUsed(Payload.User.ID, Payload.Channel.ID)
				if err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				} else if repoName == "" {
					if Payload.Callback == "job abort repo-list" {
						repoName = Payload.Actions[0].Selected[0].Value
					} else {
						if err := repoListView(false, "job abort repo-list", false); err != nil {
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

				if Payload.Callback == "job abort job-list" {
					repo, err := repoFind(Payload.Actions[0].Name)
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

					jobID, err := strconv.Atoi(Payload.Actions[0].Value)
					if err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if job, _, err := git.Jobs.GetJob(project.ID, jobID); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					} else if job.Status != "running" {
						SendError(Error.get([]int{1142}).join())
						return
					}

					job := &gitlab.Job{}
					if out, _, err := git.Jobs.CancelJob(project.ID, jobID); err == nil {
						job = out
					} else {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.ResponseDelay(slack.ResponseJSON{
						Channel: Payload.Channel.ID,
						User:    Payload.User.ID,
						Attachments: []slack.ResponseAttachment{
							slack.ResponseAttachment{
								Text:  fmt.Sprintf("Job *%s* will be aborted", job.Name),
								Color: "good",
							},
						},
					})

					return
				} else if repoName != "" {
					sended := false

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

					Slack.ResponseDelay(slack.ResponseJSON{
						User:    Payload.User.ID,
						Channel: Payload.Channel.ID,
						Text:    "> Finding pipeline of current release",
					})

					release := fmt.Sprintf("release-%d", repo.Release)

					pipelines, _, err := git.Pipelines.ListProjectPipelines(project.ID, &gitlab.ListProjectPipelinesOptions{Ref: &release})
					if err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					pid := 0

					if len(pipelines) == 0 {
						Slack.ResponseDelay(slack.ResponseJSON{
							User:    Payload.User.ID,
							Channel: Payload.Channel.ID,
							Text:    "> Could not find pipeline. Will be create new",
						})

						if pipeline, _, err := git.Pipelines.CreatePipeline(project.ID, &gitlab.CreatePipelineOptions{Ref: &release}); err == nil {
							pid = pipeline.ID
						} else {
							SendError(Error.push(500, 100, err.Error()).join())
							return
						}
					} else {
						pid = pipelines[0].ID
					}

					jobs, _, err := git.Jobs.ListPipelineJobs(project.ID, pid, &gitlab.ListJobsOptions{})
					if err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if len(jobs) == 0 {
						SendError(Error.get([]int{1140}).join())
						return
					}

					list := map[string]map[string][]string{}
					for _, job := range jobs {
						if job.Status != "running" {
							continue
						}

						if _, ok := list[job.Stage]; !ok {
							list[job.Stage] = map[string][]string{}
						}

						list[job.Stage][job.Name] = []string{job.Status, strconv.Itoa(job.ID)}
					}

					stages := []string{}
					for stage := range list {
						stages = append(stages, stage)
					}
					sort.Strings(stages)

					attachments := []slack.ResponseAttachment{}
					for _, stage := range stages {
						actions := []slack.ResponseAttachmentActions{}

						jobs := []string{}
						for job := range list[stage] {
							jobs = append(jobs, job)
						}
						sort.Strings(jobs)

						for _, job := range jobs {
							actions = append(actions, slack.ResponseAttachmentActions{
								Name:  repo.Name,
								Type:  "button",
								Text:  fmt.Sprintf("[ %+v ] %+v", list[stage][job][0], job),
								Value: list[stage][job][1],
								Style: jobStyle[list[stage][job][0]],
								Confirm: &slack.ResponseConfirm{
									Title:       fmt.Sprintf("Abort job: %s", job),
									Text:        fmt.Sprintf("Are you sure you want to abort the job \"*%s*\"? Current state: *%s*", job, list[stage][job][0]),
									TextDismiss: "No",
									TextOK:      "Yep",
								},
							})

							if len(actions) == 5 {
								attachments = append(attachments, slack.ResponseAttachment{
									Title:    stage,
									Color:    "#2eb886",
									Callback: "job abort job-list",
									Actions:  actions,
								})

								Slack.ResponseDelay(slack.ResponseJSON{
									User:        Payload.User.ID,
									Channel:     Payload.Channel.ID,
									Attachments: attachments,
								})

								attachments = []slack.ResponseAttachment{}
								actions = []slack.ResponseAttachmentActions{}

								sended = true
							}
						}

						if len(actions) > 0 {
							attachments = append(attachments, slack.ResponseAttachment{
								Title:    stage,
								Color:    "#2eb886",
								Callback: "job abort job-list",
								Actions:  actions,
							})
						}
					}

					if len(attachments) == 0 {
						if sended {
							return
						}

						attachments = append(attachments, slack.ResponseAttachment{
							Color: "#daa038",
							Text:  "A job list is empty",
						})
					}

					Slack.ResponseDelay(slack.ResponseJSON{
						User:        Payload.User.ID,
						Channel:     Payload.Channel.ID,
						Attachments: attachments,
					})

					return
				}
			}()

			return "", false
		},
	}
}

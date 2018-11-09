package server

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/dhalturin/release-manager/slack"
	gitlab "github.com/xanzy/go-gitlab"
)

func init() {
	method := "job list"

	MethodList[method] = MethodStruct{
		Help: []string{method, "run jobs from last pipeline of current release"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				repoName, err := repoUsed(Payload.User.ID, Payload.Channel.ID)
				if err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				} else if repoName == "" {
					if Payload.Callback == "job list repo-list" {
						repoName = Payload.Actions[0].Selected[0].Value
					} else {
						if err := repoListView(false, "job list repo-list", false); err != nil {
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
						jobs := []string{}
						for job := range list[stage] {
							jobs = append(jobs, job)
						}
						sort.Strings(jobs)

						fields := []slack.ResponseAttachmentField{}
						for _, job := range jobs {
							fields = append(fields, slack.ResponseAttachmentField{
								Title: job,
								Value: fmt.Sprintf("%s: <%s/-/jobs/%s|%s>", list[stage][job][0], project.WebURL, list[stage][job][1], list[stage][job][1]),
								Short: true,
							})
						}

						attachments = append(attachments, slack.ResponseAttachment{
							Text:   fmt.Sprintf("Stage: `%s`", stage),
							Fields: fields,
							Color:  "#bff442",
						})
					}

					Slack.ResponseDelay(slack.ResponseJSON{
						User:        Payload.User.ID,
						Channel:     Payload.Channel.ID,
						Attachments: attachments,
						Text:        fmt.Sprintf("> Pipeline: <%s/pipelines/%d|%d>", project.WebURL, pid, pid),
					})

					return
				}
			}()

			return "", false
		},
	}
}

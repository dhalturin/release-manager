package server

import (
	"github.com/dhalturin/release-manager/slack"
)

func init() {
	method := "repo del"

	MethodList[method] = MethodStruct{
		Help: []string{method, "removing repository"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				if Payload.Callback == "repo del value" {
					repo, err := repoFind(Payload.Actions[0].Selected[0].Value)
					if err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					if repo.Admin != Payload.User.ID {
						SendError(Error.get([]int{1121}).join())
						return
					}

					if _, err := conn.Exec("delete from repos where repo_id = $1", repo.ID); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.Response(Payload.ResponseURL, slack.ResponseJSON{Text: "", Attachments: []slack.ResponseAttachment{
						slack.ResponseAttachment{
							Color: "#bff442",
							Text:  "Repository has been successfully removed",
						},
					}})

					return
				}

				// Default action

				if err := repoListView(true, "repo del value", true); err != nil {
					SendError(Error.push(500, 100, err.Error()).join())
					return
				}
			}()

			return "", false
		},
	}
}

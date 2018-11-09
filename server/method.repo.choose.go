package server

import (
	"github.com/dhalturin/release-manager/slack"
)

func init() {
	method := "repo choose"

	MethodList[method] = MethodStruct{
		Help: []string{method, "select the repository to perform actions"},
		Func: func() (interface{}, bool) {
			go func() {
				Error := Err{Pointer: method, Meta: "method"}

				if Payload.Callback == "repo choose update" {
					if err := repoUsedUpdate(Payload.User.ID, Payload.Channel.ID, Payload.Actions[0].Selected[0].Value); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}

					Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
						Attachments: []slack.ResponseAttachment{
							slack.ResponseAttachment{
								Color: "good",
								Text:  "The repository was successfully selected",
							},
						},
					})
				} else {
					if err := repoListView(false, "repo choose update", false); err != nil {
						SendError(Error.push(500, 100, err.Error()).join())
						return
					}
				}
			}()

			return "", false
		},
	}
}

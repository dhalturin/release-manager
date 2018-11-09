package server

import (
	"regexp"
	"sort"

	"github.com/dhalturin/release-manager/slack"
)

func init() {
	method := "default"

	MethodList[method] = MethodStruct{
		Func: func() (interface{}, bool) {
			list := map[string][]string{}

			for i, j := range MethodList {
				if Payload.UserToken == "" && !j.NoToken {
					continue
				} else if Payload.UserToken != "" && j.OnlyGuest {
					continue
				}

				if len(j.Help) > 0 {
					key := regexp.MustCompile(" .*").ReplaceAllString(i, "")

					if Payload.Filter != "" {
						if key != Payload.Filter {
							continue
						}
					}

					list[key] = append(list[key], j.Help[0])
				}
			}

			if len(list) == 0 {
				return "", false
			}

			keys := []string{}
			for i := range list {
				keys = append(keys, i)
			}

			sort.Strings(keys)

			attachments := []slack.ResponseAttachment{}
			for _, j := range keys {
				actions := []slack.ResponseAttachmentActions{}

				sort.Strings(list[j])

				for _, action := range list[j] {
					actions = append(actions, slack.ResponseAttachmentActions{
						Name:  j,
						Type:  "button",
						Text:  action,
						Value: action,
						Style: "primary",
					})
				}

				attachments = append(attachments, slack.ResponseAttachment{
					// Title:    i,
					Color:    "#2eb886",
					Callback: "help show",
					Actions:  actions,
				})
			}

			Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
				Text:        "> To simplify the use of this tool, use the buttons:",
				Attachments: attachments,
			})

			return "", false
		}, NoToken: true}
}

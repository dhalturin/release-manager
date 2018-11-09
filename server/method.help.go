package server

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dhalturin/release-manager/slack"
)

func init() {
	method := "help"

	MethodList[method] = MethodStruct{
		// Help: []string{method, "view information about this tool"},
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

					list[key] = append(list[key], fmt.Sprintf("`%s %s` - %s", Payload.Command, j.Help[0], j.Help[1]))
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
				sort.Strings(list[j])

				attachments = append(attachments, slack.ResponseAttachment{
					Color: "#daa038",
					Text:  strings.Join(list[j], "\n"),
				})
			}

			Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
				Attachments: attachments,
			})

			return "", false
		}, NoToken: true}
}

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"

	// mysql driver

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	gitlab "github.com/xanzy/go-gitlab"

	"github.com/dhalturin/release-manager/data"
	"github.com/dhalturin/release-manager/slack"
)

var (
	conn *sqlx.DB
	err  error
	// Slack api
	Slack *slack.Client
	// GitLabHost string
	GitLabHost string
	// Payload PayloadStruct
	Payload PayloadStruct
)

// Run func
func Run() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		data.Config.DbUser,
		data.Config.DbPass,
		data.Config.DbHost,
		data.Config.DbPort,
		data.Config.DbName,
	)

	conn, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Printf("Initializing server")
	log.Printf("Listen: %s", data.Config.Listen)
	log.Printf("Port: %d", data.Config.Port)

	listen := fmt.Sprintf("%s:%d", data.Config.Listen, data.Config.Port)

	http.HandleFunc("/", handleFunc)

	err = http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	Error := Err{Pointer: r.URL.Path, Meta: "core"}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	post, _err := url.ParseQuery(buf.String())
	if _err != nil {
		Error.get([]int{1001}).print(w)
		return
	}

	Payload = PayloadStruct{}
	if post.Get("payload") != "" {
		if err := json.Unmarshal([]byte(post.Get("payload")), &Payload); err != nil {
			Error.push(400, 1002, err.Error()).print(w)
			return
		}
	}

	if Payload.Token == "" {
		Payload.ResponseURL = post.Get("response_url")
		Payload.Token = post.Get("token")
		Payload.Team.Domain = post.Get("team_domain")
		Payload.Team.ID = post.Get("team_id")
		Payload.Channel.ID = post.Get("channel_id")
		Payload.Channel.Name = post.Get("channel_name")
		Payload.User.Name = post.Get("user_name")
		Payload.User.ID = post.Get("user_id")
		Payload.Command = post.Get("command")
		Payload.Text = post.Get("text")
		Payload.TriggerID = post.Get("trigger_id")
	} else {
		Payload.Command = data.Command
	}

	fmt.Printf("\n-----------\nBody:\n%+v\n-----------\n", post)
	fmt.Printf("\n-----------\nPayload:\n%+v\n-----------\n\n", Payload)

	if Payload.Token != data.SlackVerification {
		Error.get([]int{1003}).print(w)
		return
	}

	if stopChannel[Payload.Channel.Name] {
		Error.get([]int{1004}).print(w)
		return
	}

	if !payloadType[Payload.Type] {
		if Payload.Command != data.Command {
			Error.get([]int{1005}).print(w)
			return
		}
	}

	user := user{}
	if err := user.find(Payload.User.ID); err != nil {
		Error.push(400, 1006, err.Error()).print(w)
		return
	}

	git := gitlab.NewClient(nil, user.Token)
	git.SetBaseURL(GitLabHost)
	if _, _, err := git.Version.GetVersion(); err == nil {
		Payload.UserToken = user.Token
	}
	
	fmt.Println(len(post))
	fmt.Printf("\n-----------\nBody:\n%+v\n-----------\n", post)
	fmt.Printf("\n-----------\nPayload:\n%+v\n-----------\n\n", Payload)

	method := Payload.Text

	if Payload.Callback != "" {
		if Payload.Callback == "help show" {
			Payload.Callback = fmt.Sprintf("%s help", Payload.Actions[0].Value)
		}

		/*
		 * method: repo del value => repo del
		 */
		method = regexp.MustCompile("(\\ [a-z-_0-9]+)$").ReplaceAllString(Payload.Callback, "")
		// method = Payload.Callback
	}

	if method == "" {
		method = "default"
	}

	if MethodList[method].Func == nil {
		Payload.Filter = method
		method = "default"
	} else {
		if Payload.UserToken == "" && !MethodList[method].NoToken {
			return
		}
	}

	Slack = slack.Init(data.SlackOAuth, data.SlackVerification)

	Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
		Attachments: []slack.ResponseAttachment{
			slack.ResponseAttachment{
				Color: "#bff442",
				Text:  "Processing request :nyancat:",
			},
		},
	})

	if method != "default" {
		if len(MethodList[method].Help) == 0 && Payload.Token == "" {
			return
		}

		// if Payload.Token != "" {
		// 	Slack.Response(Payload.ResponseURL, slack.ResponseJSON{Text: "", Attachments: []slack.ResponseAttachment{
		// 		slack.ResponseAttachment{
		// 			Color: "#bff442",
		// 			Text:  fmt.Sprintf("Loading module \"%s\". Please wait. :banana-happy:", method),
		// 		},
		// 	}})
		// }
	}

	log.Printf("method :: %+v", method)

	response, err := MethodList[method].Func()

	if err {
		SendError(response)
	} else {
		w.Write([]byte(fmt.Sprintf("%v", response)))
	}
}

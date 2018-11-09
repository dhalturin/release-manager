package server

import (
	"fmt"
	"log"
	"regexp"
	"runtime"

	"github.com/dhalturin/release-manager/data"

	"github.com/dhalturin/release-manager/slack"
)

// MethodStruct struct
type MethodStruct struct {
	Func      func() (interface{}, bool)
	NoToken   bool
	OnlyGuest bool
	Help      []string
}

// MethodInterface interface
type MethodInterface interface {
}

// MethodData struct
type MethodData struct {
	Data MethodInterface `json:"data"`
}

// MethodRelations struct
type MethodRelations struct {
	Data MethodInterface `json:"data"`
}

// MethodRelationsData struct
type MethodRelationsData struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

// MethodDataIncluded struct
type MethodDataIncluded struct {
	Data     MethodInterface `json:"data"`
	Included MethodInterface `json:"included"`
}

// MethodAttributes struct
type MethodAttributes struct {
	Attributes MethodInterface `json:"attributes"`
}

// MethodAttributesType struct
type MethodAttributesType struct {
	Type       string          `json:"type"`
	ID         int             `json:"id"`
	Attributes MethodInterface `json:"attributes"`
}

// MethodAttributesRelations struct
type MethodAttributesRelations struct {
	Type       string          `json:"type"`
	ID         int             `json:"id"`
	Attributes MethodInterface `json:"attributes"`
	Relations  MethodInterface `json:"relationships"`
}

// MethodAttributesResponse struct
type MethodAttributesResponse struct {
	Response bool `json:"response"`
}

// MethodOpts struct
type MethodOpts struct {
	UserID             string `json:"user_id"`
	UserLogin          string `json:"user_login"`
	UserPassword       string `json:"user_password"`
	UserPasswordRepeat string `json:"user_passwordRepeat"`
	UserEmail          string `json:"user_email"`
	UserCode           string `json:"user_code"`
	NodeID             int    `json:"node_id"`
	NodeProtocol       string `json:"node_protocol"`
	NodeHost           string `json:"node_host"`
	NodePort           string `json:"node_port"`
	NodePath           string `json:"node_path"`
	NodeUser           string `json:"node_user"`
	NodePassword       string `json:"node_password"`
}

// MethodList variable
var MethodList = make(map[string]MethodStruct)

// PayloadStruct struct
type PayloadStruct struct {
	Type        string            `json:"type"`
	Token       string            `json:"token"`
	TriggerID   string            `json:"trigger_id"`
	ActionTS    string            `json:"action_ts"`
	MessageTS   string            `json:"message_ts"`
	Actions     []payloadAction   `json:"actions"`
	Team        payloadTeam       `json:"team"`
	User        payloadUser       `json:"user"`
	Channel     payloadChannel    `json:"channel"`
	Submission  payloadSubmission `json:"submission"`
	Callback    string            `json:"callback_id"`
	ResponseURL string            `json:"response_url"`
	State       string            `json:"state"`
	Command     string            `json:"-"`
	Text        string            `json:"-"`
	Filter      string            `json:"-"`
	UserToken   string            `db:"user_token"`
}

// payloadAction struct
type payloadAction struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Selected []struct {
		Value string `json:"value"`
	} `json:"selected_options"`
}

type payloadTeam struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

type payloadUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type payloadChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type payloadSubmission struct {
	Channel    string `json:"channel"`
	Repository string `json:"repository"`
	Ref        string `json:"ref"`
	Token      string `json:"token"`
	Permission string `json:"permission"`
	Prefix     string `json:"prefix"`
	Release    string `json:"release"`
	Tasks      string `json:"tasks"`
}

var stopChannel = map[string]bool{
	"directmessage": true,
	"privategroup":  true,
}

var payloadType = map[string]bool{
	"dialog_submission":   true,
	"interactive_message": true,
}

var jobStyle = map[string]string{
	"success": "primary",
	"running": "primary",
	"failed":  "danger",
}

var jobSkeep = map[string]bool{
	"skipped": true,
	"created": true,
}

// var methodInteractive = map[string]bool{
// 	"help show":      true,
// 	"repo del value": true,
// }

type repo struct {
	ID         int    `db:"repo_id"`
	Name       string `db:"repo_name"`
	Ref        string `db:"repo_ref"`
	Release    int    `db:"repo_release"`
	Pipeline   int    `db:"repo_pipeline"`
	Token      string `db:"repo_token"`
	Prefix     string `db:"repo_prefix"`
	Channel    string `db:"repo_channel"`
	Permission string `db:"repo_permission"`
	Admin      string `db:"repo_admin"`
	Tasks      string `db:"repo_tasks"`
}

type tasksJSON struct {
	Task  string `json:"task"`
	Title string `json:"title"`
}

// SendError func
func SendError(response interface{}) bool {
	_, fn, line, _ := runtime.Caller(1)
	fn = regexp.MustCompile(fmt.Sprintf(".*%s/", data.Project)).ReplaceAllString(fn, "")

	OutputAttachment := []slack.ResponseAttachment{}
	OutputAttachment = append(OutputAttachment, slack.ResponseAttachment{
		Color:  "danger",
		Title:  "Module error",
		Text:   fmt.Sprintf("%+v", response),
		Footer: fmt.Sprintf("%d: %s", line, fn),
	})

	if _, err := Slack.Response(Payload.ResponseURL, slack.ResponseJSON{
		Attachments: OutputAttachment,
	}); err != nil {
		log.Print("send response error: ", err.Error())
	}

	return false
}
